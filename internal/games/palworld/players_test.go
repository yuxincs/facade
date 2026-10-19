package palworld

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlayerCount(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
		count  int
		err    string
	}{
		{name: "players", status: 200, body: `{"players":[{"name":"Alice"},{"name":"Bob"}]}`, count: 2},
		{name: "empty", status: 200, body: `{"players":[]}`},
		{name: "unauthorized", status: 401, body: "secret-password", err: "HTTP 401"},
		{name: "server error", status: 500, body: "secret-password", err: "HTTP 500"},
		{name: "redirect", status: 302, err: "HTTP 302"},
		{name: "malformed", status: 200, body: `{"players":`, err: "valid players array"},
		{name: "missing players", status: 200, body: `{}`, err: "valid players array"},
		{name: "null players", status: 200, body: `{"players":null}`, err: "valid players array"},
		{name: "wrong type", status: 200, body: `{"players":{}}`, err: "valid players array"},
		{name: "trailing data", status: 200, body: `{"players":[]} {}`},
		{name: "trailing garbage", status: 200, body: `{"players":[]} garbage`},
		{name: "exact size limit", status: 200, body: `{"players":[],"padding":"` + strings.Repeat(" ", _maxResponseSize-len(`{"players":[],"padding":""}`)) + `"}`},
		{name: "oversized JSON", status: 200, body: `{"players":[],"padding":"` + strings.Repeat(" ", _maxResponseSize) + `"}`, err: "exceeds 1 MiB"},
		{name: "oversized", status: 200, body: strings.Repeat(" ", (1<<20)+1), err: "exceeds 1 MiB"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			requests := make(chan *http.Request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r
				w.Header().Set("Location", "/unexpected")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			t.Cleanup(server.Close)
			game := playerTestGame(t, server)
			count, err := game.PlayerCount(t.Context())
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				require.NotContains(t, err.Error(), game.adminPassword)
				require.Zero(t, count)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.count, count)
			}
			r := <-requests
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "/v1/api/players", r.URL.Path)
			require.Equal(t, "application/json", r.Header.Get("Accept"))
			username, password, ok := r.BasicAuth()
			require.True(t, ok)
			require.Equal(t, "admin", username)
			require.Equal(t, game.adminPassword, password)
		})
	}
}

func TestPlayerCountContext(t *testing.T) {
	for _, name := range []string{"cancel", "deadline"} {
		t.Run(name, func(t *testing.T) {
			started := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				close(started)
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			t.Cleanup(server.Close)
			game := playerTestGame(t, server)
			ctx, cancel := context.WithCancel(t.Context())
			wantErr := context.Canceled
			if name == "deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(t.Context(), 100*time.Millisecond)
				wantErr = context.DeadlineExceeded
			}
			defer cancel()
			if name == "cancel" {
				go func() {
					<-started
					cancel()
				}()
			}
			count, err := game.PlayerCount(ctx)
			require.ErrorIs(t, err, wantErr)
			require.Zero(t, count)
		})
	}
}

func TestPlayerCountUnavailable(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	game := playerTestGame(t, server)
	server.Close()
	count, err := game.PlayerCount(t.Context())
	require.ErrorContains(t, err, "query Palworld players")
	require.Zero(t, count)
}

func playerTestGame(t *testing.T, server *httptest.Server) *Palworld {
	t.Helper()
	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	require.NoError(t, err)
	dir := t.TempDir()
	config := filepath.Join(dir, "Pal", "Saved", "Config", "LinuxServer")
	require.NoError(t, os.MkdirAll(config, 0o755))
	ini := fmt.Sprintf("[/Script/Pal.PalGameWorldSettings]\nOptionSettings=(RESTAPIEnabled=True,RESTAPIPort=%s,AdminPassword=\"secret-password\")\n", port)
	require.NoError(t, os.WriteFile(filepath.Join(config, "PalWorldSettings.ini"), []byte(ini), 0o644))
	game, err := New(dir)
	require.NoError(t, err)
	require.NotNil(t, game)
	return game
}
