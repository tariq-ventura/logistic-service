package clients_fleet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

func (c *Client) UpdateEquipmentStatus(
	ctx context.Context,
	equipmentID uuid.UUID,
	status string,
	reason string,
) error {
	payload, err := json.Marshal(updateEquipmentStatusRequest{
		Status: status,
		Reason: reason,
	})
	if err != nil {
		return fmt.Errorf("encoding equipment status: %w", err)
	}

	endpoint := fmt.Sprintf(
		"%s/api/v1/equipments/%s/status",
		c.baseURL,
		url.PathEscape(equipmentID.String()),
	)

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("creating Fleet request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("calling fleet-service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))

		return &HTTPError{
			StatusCode: response.StatusCode,
			Body:       string(body),
		}
	}

	return nil
}
