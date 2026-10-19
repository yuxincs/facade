// Facade is a game server proxy with on-demand server lifecycle management.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yuxincs/facade/internal/engine"
	"github.com/yuxincs/facade/internal/games/palworld"
)

// version can be set at build time with -ldflags '-X main.version=...'.
var version = "dev"

var _games = []func(string) (engine.Game, error){adaptConstructor(palworld.New)}

func main() {
	flags := flag.NewFlagSet("facade", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	showVersion := flags.Bool("version", false, "Print the version and exit")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Facade — on-demand game server proxy")
		_, _ = fmt.Fprintln(flags.Output(), "\nUsage: facade [flags]")
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(2)
	}
	if flags.NArg() != 0 {
		_, _ = fmt.Fprintf(os.Stderr, "facade: unexpected argument %q\n", flags.Arg(0))
		flags.Usage()
		os.Exit(2)
	}
	if *showVersion {
		_, _ = fmt.Fprintf(os.Stdout, "facade %s\n", version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dir, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "facade: get working directory: %v\n", err)
		os.Exit(1)
	}
	game, err := selectGame(ctx, dir, _games...)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "facade: %v\n", err)
		os.Exit(1)
	}
	slog.Info("detected game", "game", fmt.Sprintf("%T", game), "directory", dir)
	e := engine.Engine{Game: game}
	if err := e.Run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "facade: engine: %v\n", err)
		os.Exit(1)
	}
}

func adaptConstructor[T engine.Game](newGame func(string) (T, error)) func(string) (engine.Game, error) {
	return func(dir string) (engine.Game, error) { return newGame(dir) }
}

func selectGame(ctx context.Context, dir string, constructors ...func(string) (engine.Game, error)) (engine.Game, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var selected engine.Game
	var matches []string
	for _, newGame := range constructors {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		game, err := newGame(dir)
		if errors.Is(err, errors.ErrUnsupported) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("initialize game: %w", err)
		}
		selected = game
		matches = append(matches, fmt.Sprintf("%T", game))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no supported game detected in %q", dir)
	case 1:
		return selected, nil
	default:
		return nil, fmt.Errorf("multiple games matched %q: %s", dir, strings.Join(matches, ", "))
	}
}
