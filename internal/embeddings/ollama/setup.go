package embeddings_ollama

import (
	"net/http"
	"time"

	"github.com/tariq-ventura/logistic-service/internal/validations"
)

const (
	defaultDimensions = 768
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type embeddingRequest struct {
	Model      string `json:"model"`
	Input      string `json:"input"`
	Truncate   bool   `json:"truncate"`
	Dimensions int    `json:"dimensions"`
	KeepAlive  string `json:"keep_alive"`
}

type embeddingResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`
}

func NewClient() (*Client, error) {
	baseURL, err := validations.RequiredEnv("OLLAMA_URL")

	if err != nil {
		return nil, err
	}

	model, err := validations.RequiredEnv("OLLAMA_EMBEDDING_MODEL")

	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) Provider() string {
	return "ollama"
}

func (c *Client) Model() string {
	return c.model
}

func (c *Client) Dimensions() int {
	return defaultDimensions
}
