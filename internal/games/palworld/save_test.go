package palworld

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
)

func TestSave(t *testing.T) {
	t.Parallel()

	for _, status := range []int{200, 400, 401, 500, 302} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			t.Parallel()

			requests := make(chan *http.Request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r
				w.Header().Set("Location", "/unexpected")
				w.WriteHeader(status)
				_, _ = fmt.Fprint(w, "secret-password")
			}))
			t.Cleanup(server.Close)
			game := playerTestGame(t, server)
			require.Implements(t, (*engine.Saver)(nil), game)
			err := game.Save(t.Context())
			if status == 200 {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, fmt.Sprintf("HTTP %d", status))
				require.NotContains(t, err.Error(), game.adminPassword)
			}
			r := <-requests
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/v1/api/save", r.URL.Path)
			username, password, ok := r.BasicAuth()
			require.True(t, ok)
			require.Equal(t, "admin", username)
			require.Equal(t, game.adminPassword, password)
		})
	}
}

func TestSaveCanceled(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	game := playerTestGame(t, server)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() {
		<-started
		cancel()
	}()
	require.ErrorIs(t, game.Save(ctx), context.Canceled)
}
