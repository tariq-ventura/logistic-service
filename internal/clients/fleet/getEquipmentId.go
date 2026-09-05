package clients_fleet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (c *Client) GetEquipmentByID(
	ctx context.Context,
	equipmentID uuid.UUID,
) (*Equipment, error) {
	endpoint := fmt.Sprintf(
		"%s/api/v1/equipments/%s",
		c.baseURL,
		url.PathEscape(equipmentID.String()),
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating Fleet request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("calling fleet-service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))

		return nil, &HTTPError{
			StatusCode: response.StatusCode,
			Body:       string(body),
		}
	}

	var result equipmentResponse

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding Fleet response: %w", err)
	}

	return &result.Data, nil
}
