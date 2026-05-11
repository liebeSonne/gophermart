package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/liebeSonne/gophermart/internal/config"
	iocloser "github.com/liebeSonne/gophermart/internal/io/closer"
	ilogger "github.com/liebeSonne/gophermart/internal/logger"
)

func runApp(
	ctx context.Context,
	cfg config.Config,
	closer *iocloser.MultiCloser,
	logger ilogger.Logger,
) error {
	connection, err := newConnectionContainer(ctx, cfg, closer, logger)
	if err != nil {
		logger.Errorw("error creating connection container", "err", err)
		return err
	}

	dependency, err := newDependencyContainer(cfg, logger, connection)
	if err != nil {
		logger.Errorw("error creating dependency container", "err", err)
		return err
	}

	runProducers(ctx, dependency)
	runWorkers(ctx, logger, dependency)

	router, err := initRouter(dependency, logger)
	if err != nil {
		return err
	}

	logger.Infow("starting server")

	srv := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrorsCh := make(chan error, 1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrorsCh <- err
		}
	}()

	select {
	case err := <-serverErrorsCh:
		return err
	case <-ctx.Done():
		logger.Infow("starting server shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			srv.Close()
			return err
		}

		logger.Infow("server shutdown complete")
	}

	return nil
}
