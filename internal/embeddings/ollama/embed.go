package embeddings_ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) embed(
	ctx context.Context,
	input string,
) ([]float32, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("embedding input is empty")
	}

	body := embeddingRequest{
		Model:      c.model,
		Input:      input,
		Truncate:   true,
		Dimensions: defaultDimensions,
		KeepAlive:  "10m",
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(
			"encoding Ollama request: %w",
			err,
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/embed",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"creating Ollama request: %w",
			err,
		)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"calling Ollama: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(
			io.LimitReader(response.Body, 8*1024),
		)

		return nil, fmt.Errorf(
			"Ollama returned HTTP %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var result embeddingResponse

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"decoding Ollama response: %w",
			err,
		)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf(
			"Ollama returned no embeddings",
		)
	}

	vector := result.Embeddings[0]

	if len(vector) != defaultDimensions {
		return nil, fmt.Errorf(
			"invalid embedding dimensions: expected %d, received %d",
			defaultDimensions,
			len(vector),
		)
	}

	return vector, nil
}
