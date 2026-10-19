package proxy_test

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/proxy"
)

func TestUDPForwarding(t *testing.T) {
	t.Parallel()
	backend := udpListener(t)
	addresses := make(chan string, 16)
	echoDone := make(chan struct{})
	go func() {
		defer close(echoDone)
		buffer := make([]byte, 65535)
		for {
			n, addr, err := backend.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			addresses <- addr.String()
			if _, err := backend.WriteToUDP(buffer[:n], addr); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = backend.Close()
		<-echoDone
	})
	listener := udpListener(t)
	cancel, done := runUDP(t, &proxy.UDP{BackendAddress: backend.LocalAddr().String()}, listener)

	first := udpClient(t, listener)
	second := udpClient(t, listener)
	for _, payload := range [][]byte{[]byte("hello"), {}, bytes.Repeat([]byte("x"), 8000)} {
		udpRoundTrip(t, first, payload)
	}
	firstAddress := <-addresses
	require.Equal(t, firstAddress, <-addresses)
	require.Equal(t, firstAddress, <-addresses)
	udpRoundTrip(t, second, []byte("second client"))
	require.NotEqual(t, firstAddress, <-addresses)
	udpRoundTrip(t, first, []byte("first client again"))
	require.Equal(t, firstAddress, <-addresses)

	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(3 * time.Second):
		t.Fatal("proxy did not stop after cancellation")
	}
	_, _, err := listener.ReadFromUDP(make([]byte, 1))
	require.ErrorIs(t, err, net.ErrClosed)
}

func TestUDPIdleSession(t *testing.T) {
	t.Parallel()
	backend := udpListener(t)
	listener := udpListener(t)
	_, _ = runUDP(t, &proxy.UDP{
		BackendAddress: backend.LocalAddr().String(),
		IdleTimeout:    30 * time.Millisecond,
	}, listener)
	client := udpClient(t, listener)
	require.NoError(t, client.SetDeadline(time.Now().Add(3*time.Second)))
	_, err := client.Write([]byte("hello"))
	require.NoError(t, err)
	require.NoError(t, backend.SetReadDeadline(time.Now().Add(3*time.Second)))
	buffer := make([]byte, 32)
	_, addr, err := backend.ReadFromUDP(buffer)
	require.NoError(t, err)
	_, err = backend.WriteToUDP([]byte("hello"), addr)
	require.NoError(t, err)
	_, err = client.Read(buffer)
	require.NoError(t, err)

	// Wait beyond the idle timeout before sending to the old session.
	<-time.After(150 * time.Millisecond)
	_, err = backend.WriteToUDP([]byte("expired"), addr)
	require.NoError(t, err)
	require.NoError(t, client.SetReadDeadline(time.Now().Add(50*time.Millisecond)))
	_, err = client.Read(buffer)
	var timeout net.Error
	require.ErrorAs(t, err, &timeout)
	require.True(t, timeout.Timeout())

	_, err = client.Write([]byte("new session"))
	require.NoError(t, err)
	_, addr, err = backend.ReadFromUDP(buffer)
	require.NoError(t, err)
	_, err = backend.WriteToUDP([]byte("reconnected"), addr)
	require.NoError(t, err)
	require.NoError(t, client.SetReadDeadline(time.Now().Add(3*time.Second)))
	n, err := client.Read(buffer)
	require.NoError(t, err)
	require.Equal(t, "reconnected", string(buffer[:n]))
}

func TestUDPListenerClosed(t *testing.T) {
	t.Parallel()
	backend := udpListener(t)
	listener := udpListener(t)
	_, done := runUDP(t, &proxy.UDP{BackendAddress: backend.LocalAddr().String()}, listener)
	require.NoError(t, listener.Close())
	select {
	case err := <-done:
		require.ErrorIs(t, err, net.ErrClosed)
	case <-time.After(3 * time.Second):
		t.Fatal("proxy did not stop after listener closed")
	}
}

func TestUDPInvalidConfiguration(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		udp  proxy.UDP
		err  string
	}{
		{name: "negative timeout", udp: proxy.UDP{IdleTimeout: -time.Second}, err: "idle timeout"},
		{name: "invalid backend", udp: proxy.UDP{BackendAddress: "invalid"}, err: "resolve UDP backend"},
		{name: "zero port", udp: proxy.UDP{BackendAddress: "127.0.0.1:0"}, err: "port must not be zero"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorContains(t, tt.udp.Run(t.Context(), udpListener(t)), tt.err)
		})
	}
}

func udpListener(t *testing.T) *net.UDPConn {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func udpClient(t *testing.T, listener *net.UDPConn) *net.UDPConn {
	t.Helper()
	conn, err := net.DialUDP("udp4", nil, listener.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func udpRoundTrip(t *testing.T, conn *net.UDPConn, payload []byte) {
	t.Helper()
	require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
	_, err := conn.Write(payload)
	require.NoError(t, err)
	buffer := make([]byte, 65535)
	n, err := conn.Read(buffer)
	require.NoError(t, err)
	require.Equal(t, payload, buffer[:n])
}

func runUDP(t *testing.T, p *proxy.UDP, listener *net.UDPConn) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- p.Run(ctx, listener) }()
	t.Cleanup(cancel)
	return cancel, done
}
