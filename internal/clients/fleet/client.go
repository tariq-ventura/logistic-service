package clients_fleet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) ListAvailableEquipments(
	ctx context.Context,
	equipmentType string,
) ([]Equipment, error) {
	const pageSize = 100

	page := 1
	equipments := make([]Equipment, 0)

	for {
		response, err := c.listEquipments(
			ctx,
			equipmentType,
			"AVAILABLE",
			page,
			pageSize,
		)
		if err != nil {
			return nil, err
		}

		equipments = append(equipments, response.Data...)

		if response.Pagination.TotalPages == 0 ||
			page >= response.Pagination.TotalPages {
			break
		}

		page++
	}

	return equipments, nil
}

func (c *Client) listEquipments(
	ctx context.Context,
	equipmentType string,
	status string,
	page int,
	pageSize int,
) (*ListEquipmentsResponse, error) {
	endpoint, err := url.Parse(c.baseURL + "/api/v1/equipments")
	if err != nil {
		return nil, fmt.Errorf("building fleet URL: %w", err)
	}

	query := endpoint.Query()
	query.Set("type", strings.ToUpper(strings.TrimSpace(equipmentType)))
	query.Set("status", strings.ToUpper(strings.TrimSpace(status)))
	query.Set("page", strconv.Itoa(page))
	query.Set("pageSize", strconv.Itoa(pageSize))

	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating fleet request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("calling fleet-service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))

		return nil, fmt.Errorf(
			"fleet-service returned HTTP %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var result ListEquipmentsResponse

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding fleet response: %w", err)
	}

	return &result, nil
}
