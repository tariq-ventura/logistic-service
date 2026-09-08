package embeddings

import (
	"context"
	"errors"

	embeddings_ollama "github.com/tariq-ventura/logistic-service/internal/embeddings/ollama"
	embeddings_vertex "github.com/tariq-ventura/logistic-service/internal/embeddings/vertex"
	"github.com/tariq-ventura/logistic-service/internal/logging"
	"github.com/tariq-ventura/logistic-service/internal/validations"
)

type IClient interface {
	EmbedDocument(ctx context.Context, title string, text string) ([]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
	Provider() string
	Model() string
	Dimensions() int
}

var SetupEmbedding = func(ctx context.Context, l logging.ILogging) (IClient, error) {
	embeddingType, err := validations.RequiredEnv("EMBEDDING_CONTEXT")
	if err != nil {
		return nil, err
	}

	l.LogInfo("Embedding client selected", map[string]any{"embedding_type": embeddingType})

	switch embeddingType {
	case "vertex":
		vertex, err := embeddings_vertex.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		return vertex, nil
	case "ollama":
		ollama, err := embeddings_ollama.NewClient()
		if err != nil {
			return nil, err
		}
		return ollama, nil
	default:
		return nil, errors.New("unsupported embedding client")
	}
}
