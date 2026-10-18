// Facade is a game server proxy with on-demand server lifecycle management.
package main

import (
	"fmt"
	"io"
	"os"
)

// version can be set at build time with -ldflags '-X main.version=...'.
var version = "dev"

const usage = `Facade — on-demand game server proxy

Usage:
  facade help       Show this help
  facade version    Print the version

Proxying and server lifecycle management are under development.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if len(args) != 1 {
		fmt.Fprintln(stderr, "facade: expected a single command; use 'facade help'")
		return 2
	}

	switch args[0] {
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
	case "version", "--version":
		fmt.Fprintf(stdout, "facade %s\n", version)
	default:
		fmt.Fprintf(stderr, "facade: unknown command %q; use 'facade help'\n", args[0])
		return 2
	}
	return 0
}
