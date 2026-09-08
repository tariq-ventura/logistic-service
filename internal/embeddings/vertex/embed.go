package embeddings_vertex

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
	content string,
	taskType string,
	title string,
) ([]float32, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("embedding content is empty")
	}

	body := embeddingRequest{
		Instances: []embeddingInstance{
			{
				Content:  content,
				TaskType: taskType,
				Title:    title,
			},
		},
		Parameters: embeddingParameters{
			AutoTruncate: true,
		},
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf(
			"encoding embedding request: %w",
			err,
		)
	}

	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		c.location,
		c.projectID,
		c.location,
		c.model,
	)

	token, err := c.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf(
			"getting Google access token: %w",
			err,
		)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"creating embedding request: %w",
			err,
		)
	}

	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"calling Vertex AI: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(
			io.LimitReader(response.Body, 8*1024),
		)

		return nil, fmt.Errorf(
			"Vertex AI returned HTTP %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(responseBody)),
		)
	}

	var result embeddingResponse

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"decoding Vertex AI response: %w",
			err,
		)
	}

	if len(result.Predictions) == 0 {
		return nil, fmt.Errorf(
			"Vertex AI returned no predictions",
		)
	}

	vector := result.Predictions[0].Embeddings.Values

	if len(vector) != vectorSize {
		return nil, fmt.Errorf(
			"invalid embedding dimensions: expected %d, received %d",
			vectorSize,
			len(vector),
		)
	}

	return vector, nil
}
