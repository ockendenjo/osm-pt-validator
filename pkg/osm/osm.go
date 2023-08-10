package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-xray-sdk-go/xray"
	"io"
	"net/http"
	"sync"
	"time"
)

const defaultBaseUrl = "https://api.openstreetmap.org/api/0.6"

func NewClient() *OSMClient {
	httpClient := http.Client{Timeout: time.Duration(3) * time.Second}
	return &OSMClient{
		httpClient: httpClient,
		baseUrl:    defaultBaseUrl,
		nodeCache:  NodeCache{v: map[int64]Node{}},
		wayCache:   WayCache{v: map[int64]Way{}},
	}
}

type OSMClient struct {
	httpClient http.Client
	baseUrl    string
	nodeCache  NodeCache
	wayCache   WayCache
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

type WayCache struct {
	mu sync.Mutex
	v  map[int64]Way
}

func (c *OSMClient) GetWay(ctx context.Context, wayId int64) (Way, error) {
	cacheWay, found := c.getCachedWay(wayId)
	if found {
		return cacheWay, nil
	}

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

	c.cacheWay(way)
	return way, nil
}

type NodeCache struct {
	mu sync.Mutex
	v  map[int64]Node
}

func (c *OSMClient) GetNode(ctx context.Context, nodeId int64) (Node, error) {
	cacheNode, found := c.getCachedNode(nodeId)
	if found {
		return cacheNode, nil
	}

	url := fmt.Sprintf("%s/node/%d.json", c.baseUrl, nodeId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Node{}, err
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Node{}, err
	}

	if response.StatusCode != 200 {
		return Node{}, fmt.Errorf("HTTP status code %d", response.StatusCode)
	}

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Node{}, err
	}

	var node Node
	err = json.Unmarshal(bytes, &node)
	if err != nil {
		return Node{}, err
	}

	c.cacheNode(node)
	return node, nil
}

func (c *OSMClient) getCachedNode(nodeId int64) (Node, bool) {
	c.nodeCache.mu.Lock()
	defer c.nodeCache.mu.Unlock()
	node, found := c.nodeCache.v[nodeId]
	return node, found
}

func (c *OSMClient) cacheNode(node Node) {
	nodeId := node.Elements[0].ID
	c.nodeCache.mu.Lock()
	defer c.nodeCache.mu.Unlock()
	c.nodeCache.v[nodeId] = node
}

func (c *OSMClient) getCachedWay(wayId int64) (Way, bool) {
	c.wayCache.mu.Lock()
	defer c.wayCache.mu.Unlock()
	way, found := c.wayCache.v[wayId]
	return way, found
}

func (c *OSMClient) cacheWay(way Way) {
	wayId := way.Elements[0].ID
	c.wayCache.mu.Lock()
	defer c.wayCache.mu.Unlock()
	c.wayCache.v[wayId] = way
}
