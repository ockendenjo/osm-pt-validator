package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	"io"
	"net/http"
	"time"
)

const defaultBaseUrl = "https://api.openstreetmap.org/api/0.6"

func NewClient() *OSMClient {
	httpClient := http.Client{Timeout: time.Duration(3) * time.Second}
	return &OSMClient{httpClient: httpClient, baseUrl: defaultBaseUrl}
}

type OSMClient struct {
	httpClient http.Client
	baseUrl    string
}

func (c *OSMClient) WithBaseUrl(baseUrl string) *OSMClient {
	c.baseUrl = baseUrl
	return c
}

func (c *OSMClient) WithXRay() *OSMClient {
	c.httpClient.Transport = xray.RoundTripper(http.DefaultTransport)
	return c
}

func (c *OSMClient) GetRelation(ctx context.Context, relationId int64) (Relation, error) {
	url := fmt.Sprintf("%s/relation/%d.json", c.baseUrl, relationId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Relation{}, err
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Relation{}, err
	}

	if response.StatusCode != 200 {
		return Relation{}, fmt.Errorf("HTTP status code %d", response.StatusCode)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Relation{}, err
	}

	var relation Relation
	err = json.Unmarshal(bytes, &relation)
	if err != nil {
		return Relation{}, err
	}
	return relation, nil
}

func (c *OSMClient) GetWay(ctx context.Context, wayId int64) (Way, error) {
	url := fmt.Sprintf("%s/way/%d.json", c.baseUrl, wayId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Way{}, err
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Way{}, err
	}

	if response.StatusCode != 200 {
		return Way{}, fmt.Errorf("HTTP status code %d", response.StatusCode)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Way{}, err
	}

	var way Way
	err = json.Unmarshal(bytes, &way)
	if err != nil {
		return Way{}, err
	}
	return way, nil
}
