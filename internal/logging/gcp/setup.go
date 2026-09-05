package logging_gcp

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/logging"
)

type GCPClient struct {
	logs   *logging.Logger
	errors *errorreporting.Client
	ctx    context.Context
}

var NewGCPClient = func(ctx context.Context) (*GCPClient, error) {
	projectID, find := os.LookupEnv("GCP_PROJECT_ID")

	if !find {
		return nil, fmt.Errorf("GCP_PROJECT_ID not found in environment variables")
	}

	logs, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	logger := logs.Logger("candystore")

	errors, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName:    "candystore",
		ServiceVersion: "0.0.0",
		OnError: func(err error) {
			log.Printf("Could not report the error: %v", err)
		},
	})

	if err != nil {
		return nil, err
	}

	return &GCPClient{
		logs:   logger,
		errors: errors,
		ctx:    ctx,
	}, nil
}
