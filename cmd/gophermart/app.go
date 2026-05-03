package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/config"
)

func runApp(
	ctx context.Context,
	cfg config.Config,
	logger *logrus.Logger,
) error {
	router, err := initRouter()
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
