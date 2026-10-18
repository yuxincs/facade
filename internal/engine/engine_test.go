package engine_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuxincs/facade/internal/engine"
)

type adapterFunc func(string) (bool, error)

func (f adapterFunc) Detect(dir string) (bool, error) { return f(dir) }

func (adapterFunc) Protocol() string { return "tcp" }

func TestRun(t *testing.T) {
	t.Parallel()
	dir, err := os.Getwd()
	require.NoError(t, err)
	failure := errors.New("cannot inspect installation")
	match := adapterFunc(func(got string) (bool, error) {
		return got == dir, nil
	})
	miss := adapterFunc(func(string) (bool, error) { return false, nil })
	broken := adapterFunc(func(string) (bool, error) { return false, failure })

	for _, test := range []struct {
		name     string
		adapters []engine.Adapter
		wantErr  string
		cause    error
	}{
		{name: "no adapters", wantErr: "no supported game detected"},
		{name: "no match", adapters: []engine.Adapter{miss}, wantErr: "no supported game detected"},
		{name: "one match", adapters: []engine.Adapter{miss, match, miss}},
		{name: "multiple matches", adapters: []engine.Adapter{match, match}, wantErr: "multiple game adapters matched"},
		{name: "detection error", adapters: []engine.Adapter{broken}, wantErr: "detect engine_test.adapterFunc", cause: failure},
		{name: "error after match", adapters: []engine.Adapter{match, broken}, wantErr: "detect engine_test.adapterFunc", cause: failure},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			e := engine.Engine{Adapters: test.adapters}
			err := e.Run(t.Context())
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
			if test.cause != nil {
				require.ErrorIs(t, err, test.cause)
			}
		})
	}
}

func TestRunCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	e := engine.Engine{Adapters: []engine.Adapter{
		adapterFunc(func(string) (bool, error) {
			t.Fatal("Detect called with canceled context")
			return false, nil
		}),
	}}
	require.ErrorIs(t, e.Run(ctx), context.Canceled)
}
