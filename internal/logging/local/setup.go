package logging_local

import (
	"context"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

type LocalClient struct {
	logger *log.Logger
	ctx    context.Context
}

var NewLocalClient = func(ctx context.Context) (*LocalClient, error) {
	logger := log.New()

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("No se pudo abrir archivo de logs:", err)
		return nil, err
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	logger.SetOutput(multiWriter)

	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(log.DebugLevel)

	return &LocalClient{
		logger: logger,
		ctx:    ctx,
	}, nil
}
