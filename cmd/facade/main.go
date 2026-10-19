// Facade is a game server proxy with on-demand server lifecycle management.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuxincs/facade/internal/engine"
	"github.com/yuxincs/facade/internal/games/palworld"
)

// version can be set at build time with -ldflags '-X main.version=...'.
var version = "dev"

var _games = []engine.Game{&palworld.Palworld{}}

func main() {
	flags := flag.NewFlagSet("facade", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	showVersion := flags.Bool("version", false, "Print the version and exit")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Facade — on-demand game server proxy")
		fmt.Fprintln(flags.Output(), "\nUsage: facade [flags]")
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		os.Exit(2)
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "facade: unexpected argument %q\n", flags.Arg(0))
		flags.Usage()
		os.Exit(2)
	}
	if *showVersion {
		fmt.Fprintf(os.Stdout, "facade %s\n", version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	e := engine.Engine{
		Games: _games,
	}
	if err := e.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "facade: engine: %v\n", err)
		os.Exit(1)
	}
}
