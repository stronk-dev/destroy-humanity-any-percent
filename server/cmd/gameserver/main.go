package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/gameserver"
	"cloud-clicker/server/save"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("gameserver stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseURL := os.Getenv("DATABASE_URL")
	serverID := os.Getenv("CLOUD_CLICKER_SERVER_ID")
	key, err := base64.StdEncoding.DecodeString(os.Getenv("CLOUD_CLICKER_JWT_KEY"))
	if err != nil || databaseURL == "" || serverID == "" || len(key) < 32 {
		return gameserver.ErrComposition
	}
	repositoryRoot := os.Getenv("CLOUD_CLICKER_REPOSITORY_ROOT")
	if repositoryRoot == "" {
		repositoryRoot = "."
	}
	activity := os.Getenv("CLOUD_CLICKER_ACTIVITY_BRACKET")
	if activity == "" {
		activity = "activity.standard"
	}
	address := os.Getenv("LISTEN_ADDR")
	if address == "" {
		address = ":8080"
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	composition, err := gameserver.Compose(ctx, gameserver.CompositionConfig{DB: db, RepositoryRoot: repositoryRoot, ServerID: serverID,
		ActivityBracket: activity, SigningKeys: account.SigningKeys{CurrentID: "runtime", Current: key}, Logger: logger})
	if err != nil {
		return err
	}
	if err := composition.Server.Start(ctx); err != nil {
		return err
	}
	httpServer := &http.Server{Addr: address, Handler: composition.Server.Handler(), ReadHeaderTimeout: 5 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- httpServer.ListenAndServe() }()
	listenerErr, runtimeErr := waitForStop(ctx, serveErr, composition.Server.Failures())
	drainCtx, cancel := context.WithTimeout(context.Background(), composition.Node.DrainTimeout())
	defer cancel()
	drainErr := composition.Server.Drain(drainCtx, time.Now().UTC())
	shutdownErr := httpServer.Shutdown(drainCtx)
	return errors.Join(listenerErr, runtimeErr, drainErr, shutdownErr)
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
