package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
)

const defaultBaseUrl = "https://api.openstreetmap.org/api/0.6"
const defaultParallelReqs = 2

func NewClient(userAgent string) *OSMClient {
	httpClient := http.Client{Timeout: time.Duration(3) * time.Second}
	return &OSMClient{
		httpClient:   httpClient,
		baseUrl:      defaultBaseUrl,
		nodeCache:    NodeCache{v: map[int64]Node{}},
		wayCache:     WayCache{v: map[int64]Way{}},
		userAgent:    userAgent,
		parallelReqs: defaultParallelReqs,
	}
}

type OSMClient struct {
	httpClient   http.Client
	baseUrl      string
	nodeCache    NodeCache
	wayCache     WayCache
	userAgent    string
	parallelReqs int
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
	req.Header.Set("User-Agent", c.userAgent)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Relation{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Relation{}, err
	}

	if response.StatusCode != 200 {
		return Relation{}, HttpStatusError{response.StatusCode, string(bytes)}
	}

	var relation relationsResponse
	err = json.Unmarshal(bytes, &relation)
	if err != nil {
		return Relation{}, err
	}
	return relation.Elements[0], nil
}

func (c *OSMClient) GetRelationRelations(ctx context.Context, relationId int64) ([]Relation, error) {
	url := fmt.Sprintf("%s/relation/%d/relations.json", c.baseUrl, relationId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, HttpStatusError{response.StatusCode, string(bytes)}
	}

	var relation relationsResponse
	err = json.Unmarshal(bytes, &relation)
	if err != nil {
		return nil, err
	}
	return relation.Elements, nil

}

type relationsResponse struct {
	Elements []Relation `json:"elements"`
}

type WayCache struct {
	mu sync.Mutex
	v  map[int64]Way
}

type wayResponse struct {
	Elements []Way `json:"elements"`
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
	req.Header.Set("User-Agent", c.userAgent)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Way{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Way{}, err
	}

	if response.StatusCode != 200 {
		return Way{}, HttpStatusError{response.StatusCode, string(bytes)}
	}

	var wayRes wayResponse
	err = json.Unmarshal(bytes, &wayRes)
	if err != nil {
		return Way{}, err
	}

	way := wayRes.Elements[0]
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
	req.Header.Set("User-Agent", c.userAgent)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return Node{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return Node{}, err
	}

	if response.StatusCode != 200 {
		return Node{}, HttpStatusError{response.StatusCode, string(bytes)}
	}

	var nodeRes nodeResponse
	err = json.Unmarshal(bytes, &nodeRes)
	if err != nil {
		return Node{}, err
	}

	node := nodeRes.Elements[0]
	c.cacheNode(node)
	return node, nil
}

type nodeResponse struct {
	Elements []Node `json:"elements"`
}

func (c *OSMClient) getCachedNode(nodeId int64) (Node, bool) {
	c.nodeCache.mu.Lock()
	defer c.nodeCache.mu.Unlock()
	node, found := c.nodeCache.v[nodeId]
	return node, found
}

func (c *OSMClient) cacheNode(node Node) {
	nodeId := node.ID
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
	wayId := way.ID
	c.wayCache.mu.Lock()
	defer c.wayCache.mu.Unlock()
	c.wayCache.v[wayId] = way
}

type HttpStatusError struct {
	StatusCode   int
	ResponseBody string
}

func (e HttpStatusError) Error() string {
	return fmt.Sprintf("HTTP status code %d", e.StatusCode)
}
