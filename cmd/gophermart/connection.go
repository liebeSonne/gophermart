package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	apiclient "github.com/liebeSonne/gophermart/api/client"
	"github.com/liebeSonne/gophermart/internal/config"
	iocloser "github.com/liebeSonne/gophermart/internal/io/closer"
	"github.com/liebeSonne/gophermart/internal/repository/database"
)

type connectionContainer struct {
	DBClient      database.Client
	AccrualClient apiclient.ClientWithResponsesInterface
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

	accrualAPIClient, err := apiclient.NewClientWithResponses(cfg.AccrualSystemAddress)
	if err != nil {
		return nil, fmt.Errorf("unable to create accrual api client: %w", err)
	}

	return &connectionContainer{
		DBClient:      client,
		AccrualClient: accrualAPIClient,
	}, nil
}
