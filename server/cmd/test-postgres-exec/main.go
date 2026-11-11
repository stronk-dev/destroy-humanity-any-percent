// test-postgres-exec is a local/CI go test -exec adapter, never a game service.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	os.Exit(newExecutor(os.Stdin, os.Stdout, os.Stderr).run(ctx, os.Args[1:], os.Environ()))
}
