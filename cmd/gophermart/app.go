package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/config"
	iocloser "github.com/liebeSonne/gophermart/internal/io/closer"
)

func runApp(
	ctx context.Context,
	cfg config.Config,
	closer *iocloser.MultiCloser,
	logger *logrus.Logger,
) error {
	connection, err := newConnectionContainer(ctx, cfg, closer, logger)
	if err != nil {
		logger.WithError(err).Error("error creating connection container")
		return err
	}

	dependency, err := newDependencyContainer(cfg, connection)
	if err != nil {
		logger.WithError(err).Error("error creating dependency container")
		return err
	}

	router, err := initRouter(dependency, logger)
	if err != nil {
		return err
	}

	logger.Infoln("starting server")

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
		logger.Infoln("starting server shutdown")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			srv.Close()
			return err
		}

		logger.Infoln("server shutdown complete")
	}

	return nil
}
