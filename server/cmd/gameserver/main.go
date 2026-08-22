package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/deploymentconfig"
	"cloud-clicker/server/gameserver"
	"cloud-clicker/server/save"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if len(os.Args) == 2 && os.Args[1] == "validate-config" {
		config, err := deploymentconfig.LoadEnvironment()
		if err != nil {
			logger.Error("deployment configuration invalid", "error", err)
			os.Exit(1)
		}
		logger.Info("deployment configuration valid", "mode", config.Mode)
		return
	}
	if len(os.Args) != 1 {
		logger.Error("gameserver stopped", "error", "usage: gameserver [validate-config]")
		os.Exit(1)
	}
	if err := run(logger); err != nil {
		logger.Error("gameserver stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	runtime, err := deploymentconfig.LoadEnvironment()
	if err != nil {
		return err
	}
	ctx, stop := shutdownContext(context.Background())
	defer stop()
	db, err := save.OpenPostgres(ctx, runtime.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	signingKeys, bootstrapKeys := compositionKeys(runtime)
	composition, err := gameserver.Compose(ctx, gameserver.CompositionConfig{
		DB: db, RepositoryRoot: runtime.ContentRoot, ServerID: runtime.ServerID, ActivityBracket: runtime.ActivityBracket,
		PublicOrigin: runtime.PublicOrigin, TrustedProxyHops: runtime.TrustedProxyHops,
		SigningKeys: signingKeys, BootstrapKeys: bootstrapKeys, Logger: logger,
	})
	if err != nil {
		return err
	}
	if err := composition.Server.Start(ctx); err != nil {
		return err
	}
	httpServer := &http.Server{Addr: runtime.ListenAddress, Handler: composition.Server.Handler(), ReadHeaderTimeout: 5 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.ListenAndServe() }()
	listenerErr, runtimeErr := waitForStop(ctx, serveErr, composition.Server.Failures())
	drainCtx, cancel := context.WithTimeout(context.Background(), composition.Node.DrainTimeout())
	defer cancel()
	drainErr := composition.Server.Drain(drainCtx, time.Now().UTC())
	shutdownErr := httpServer.Shutdown(drainCtx)
	return errors.Join(listenerErr, runtimeErr, drainErr, shutdownErr)
}

func compositionKeys(runtime deploymentconfig.Config) (account.SigningKeys, account.BootstrapReceiptKeys) {
	bootstrapPrevious := map[string][]byte{}
	if runtime.Bootstrap.PreviousID != "" {
		bootstrapPrevious[runtime.Bootstrap.PreviousID] = runtime.Bootstrap.Previous
	}
	return account.SigningKeys{CurrentID: runtime.JWT.CurrentID, Current: runtime.JWT.Current,
			PreviousID: runtime.JWT.PreviousID, Previous: runtime.JWT.Previous},
		account.BootstrapReceiptKeys{CurrentID: runtime.Bootstrap.CurrentID,
			Current: runtime.Bootstrap.Current, Previous: bootstrapPrevious}
}

func shutdownContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
}

func waitForStop(ctx context.Context, serveErr, runtimeFailures <-chan error) (error, error) {
	var listenerErr, runtimeErr error
	select {
	case <-ctx.Done():
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			listenerErr = err
		}
	case runtimeErr = <-runtimeFailures:
	}
	return listenerErr, runtimeErr
}
