// Command envarium is the local development command center entry point.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"envarium.dev/envarium/internal/transport/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return cli.Main(ctx)
}
