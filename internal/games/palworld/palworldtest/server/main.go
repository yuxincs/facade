// The fake Palworld server implements only the behavior needed by Facade tests.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

func main() {
	port := flag.Uint("port", 8211, "Gameplay UDP port")
	flag.Parse()
	if *port > 65535 {
		_, _ = fmt.Fprintln(os.Stderr, "port must be between 0 and 65535")
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, uint16(*port)); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, port uint16) error {
	configs, err := filepath.Glob("Pal/Saved/Config/*/PalWorldSettings.ini")
	if err != nil {
		return err
	}
	if len(configs) != 1 {
		return errors.New("expected one settings fixture")
	}
	settings, err := os.ReadFile(configs[0])
	if err != nil {
		return err
	}
	// Install supplies this controlled fixture; this is not a general INI parser.
	restPort := regexp.MustCompile(`RESTAPIPort=(\d+)`).FindSubmatch(settings)
	password := regexp.MustCompile(`AdminPassword=("(?:[^"\\]|\\.)*")`).FindSubmatch(settings)
	if len(restPort) != 2 || len(password) != 2 {
		return errors.New("missing REST settings in fixture")
	}
	adminPassword, err := strconv.Unquote(string(password[1]))
	if err != nil {
		return errors.New("invalid password in fixture")
	}
	udp, err := net.ListenPacket("udp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	defer func() { _ = udp.Close() }()
	api, err := net.Listen("tcp4", "127.0.0.1:"+string(restPort[1]))
	if err != nil {
		return err
	}
	defer func() { _ = api.Close() }()
	events, err := os.OpenFile("events.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = events.Close() }()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/api/players", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"players":[]}`)
	})
	mux.HandleFunc("POST /v1/api/save", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(events, "saved"); err != nil {
			http.Error(w, "could not record save", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if !ok || username != "admin" || password != adminPassword {
				w.Header().Set("WWW-Authenticate", `Basic realm="Palworld"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			mux.ServeHTTP(w, r)
		}),
	}
	defer func() { _ = server.Close() }()
	if _, err := fmt.Fprintln(events, "started"); err != nil {
		return err
	}
	apiDone, udpDone := make(chan struct{}), make(chan struct{})
	var apiErr, udpErr error
	go func() { apiErr = server.Serve(api); close(apiDone) }()
	go func() { udpErr = echo(udp); close(udpDone) }()
	_, readyErr := fmt.Fprintln(os.Stdout, udp.LocalAddr())
	var runErr error
	if readyErr == nil {
		select {
		case <-ctx.Done():
		case <-apiDone:
			runErr = apiErr
		case <-udpDone:
			runErr = udpErr
		}
	}
	_ = udp.Close()
	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	shutdownErr := server.Shutdown(cleanup)
	if shutdownErr != nil {
		_ = server.Close()
	}
	<-apiDone
	<-udpDone
	_, stoppedErr := fmt.Fprintln(events, "stopped")
	return errors.Join(runErr, readyErr, shutdownErr, stoppedErr)
}

func echo(conn net.PacketConn) error {
	buffer := make([]byte, 65535)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		if _, err := conn.WriteTo(buffer[:n], addr); err != nil {
			return err
		}
	}
}
