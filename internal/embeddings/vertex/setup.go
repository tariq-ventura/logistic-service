package embeddings_vertex

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tariq-ventura/logistic-service/internal/validations"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	defaultLocation = "us-central1"
	defaultModel    = "text-multilingual-embedding-002"
	vectorSize      = 768
	cloudScope      = "https://www.googleapis.com/auth/cloud-platform"
)

type Client struct {
	projectID   string
	location    string
	model       string
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
}

type embeddingRequest struct {
	Instances  []embeddingInstance `json:"instances"`
	Parameters embeddingParameters `json:"parameters"`
}

type embeddingInstance struct {
	Content  string `json:"content"`
	TaskType string `json:"task_type"`
	Title    string `json:"title,omitempty"`
}

type embeddingParameters struct {
	AutoTruncate bool `json:"autoTruncate"`
}

type embeddingResponse struct {
	Predictions []struct {
		Embeddings struct {
			Values []float32 `json:"values"`
		} `json:"embeddings"`
	} `json:"predictions"`
}

func NewClient(ctx context.Context) (*Client, error) {
	projectID, err := validations.RequiredEnv("GCP_PROJECT_ID")
	if err != nil {
		return nil, err
	}

	location, err := validations.RequiredEnv("VERTEX_LOCATION")
	if err != nil {
		return nil, err
	}

	model, err := validations.RequiredEnv("EMBEDDING_MODEL")
	if err != nil {
		return nil, err
	}

	tokenSource, err := google.DefaultTokenSource(ctx, cloudScope)
	if err != nil {
		return nil, fmt.Errorf(
			"creating Google token source: %w",
			err,
		)
	}

	return &Client{
		projectID:   projectID,
		location:    location,
		model:       model,
		tokenSource: oauth2.ReuseTokenSource(nil, tokenSource),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

func (c *Client) Provider() string {
	return "vertex"
}

func (c *Client) Model() string {
	return c.model
}

func (c *Client) Dimensions() int {
	return vectorSize
}
