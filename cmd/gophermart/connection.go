package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/liebeSonne/gophermart/internal/config"
	iocloser "github.com/liebeSonne/gophermart/internal/io/closer"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type connectionContainer struct {
	DBClient database.Client
}

func newConnectionContainer(
	ctx context.Context,
	cfg config.Config,
	closer *iocloser.MultiCloser,
	logger *logrus.Logger,
) (*connectionContainer, error) {
	client, err := database.NewClient(ctx, cfg.DatabaseURI)
	if err != nil {
		logger.WithError(err).Error("unable to connect to database")
		return nil, fmt.Errorf("error init db client: %w", err)
	}

	if closer != nil {
		closer.AddCloser(iocloser.Func(
			func() error {
				return client.Close()
			},
		))
	}

	return &connectionContainer{
		DBClient: client,
	}, nil
}
