package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mandyl/mcp-service/internal/model"
)

// Client wraps the official Elasticsearch Go client.
type Client struct {
	es *elasticsearch.Client
}

// New creates a new ES Client connecting to the given address.
func New(addr string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{addr},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating ES client: %w", err)
	}
	return &Client{es: es}, nil
}

// Ping checks whether Elasticsearch is reachable.
func (c *Client) Ping(ctx context.Context) bool {
	res, err := c.es.Ping(c.es.Ping.WithContext(ctx))
	if err != nil || res.IsError() {
		return false
	}
	_ = res.Body.Close()
	return true
}

// Search executes a full-text search on the given index.
func (c *Client) Search(ctx context.Context, req *model.SearchRequest) (*model.SearchData, int, error) {
	// Clamp size to [1, 100].
	size := req.Size
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		return nil, model.CodeBadRequest, fmt.Errorf("size exceeds maximum (100)")
	}

	// Build the query.
	query := buildQuery(req)
	body, err := json.Marshal(query)
	if err != nil {
		return nil, model.CodeInternalError, fmt.Errorf("marshaling query: %w", err)
	}

	res, err := c.es.Search(
		c.es.Search.WithContext(ctx),
		c.es.Search.WithIndex(req.Index),
		c.es.Search.WithBody(bytes.NewReader(body)),
		c.es.Search.WithFrom(req.From),
		c.es.Search.WithSize(size),
		c.es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, model.CodeESConnectFailed, fmt.Errorf("ES search request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		raw, _ := io.ReadAll(res.Body)
		errMsg := string(raw)
		if res.StatusCode == 404 || strings.Contains(errMsg, "index_not_found_exception") {
			return nil, model.CodeIndexNotFound, fmt.Errorf("index %q not found", req.Index)
		}
		if strings.Contains(errMsg, "parsing_exception") || strings.Contains(errMsg, "search_phase_execution_exception") {
			return nil, model.CodeQuerySyntaxError, fmt.Errorf("query syntax error: %s", errMsg)
		}
		return nil, model.CodeInternalError, fmt.Errorf("ES error %d: %s", res.StatusCode, errMsg)
	}

	var esResp esSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, model.CodeInternalError, fmt.Errorf("decoding ES response: %w", err)
	}

	hits := make([]model.Hit, 0, len(esResp.Hits.Hits))
	for _, h := range esResp.Hits.Hits {
		hits = append(hits, model.Hit{
			ID:     h.ID,
			Score:  h.Score,
			Source: h.Source,
		})
	}

	return &model.SearchData{
		Total: esResp.Hits.Total.Value,
		Hits:  hits,
		From:  req.From,
		Size:  size,
	}, model.CodeSuccess, nil
}

// ListIndices returns all available index names.
func (c *Client) ListIndices(ctx context.Context) ([]string, error) {
	res, err := c.es.Cat.Indices(
		c.es.Cat.Indices.WithContext(ctx),
		c.es.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		return nil, fmt.Errorf("listing ES indices: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES indices error %d", res.StatusCode)
	}

	var indices []struct {
		Index string `json:"index"`
	}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, fmt.Errorf("decoding indices response: %w", err)
	}

	names := make([]string, 0, len(indices))
	for _, idx := range indices {
		if !strings.HasPrefix(idx.Index, ".") { // skip system indices
			names = append(names, idx.Index)
		}
	}
	return names, nil
}

// buildQuery constructs an Elasticsearch query from a SearchRequest.
func buildQuery(req *model.SearchRequest) map[string]interface{} {
	must := []interface{}{
		map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"*"},
			},
		},
	}

	if req.Filters != nil && req.Filters.Field != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				req.Filters.Field: req.Filters.Value,
			},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
			},
		},
	}
}

// --- internal ES response types ---

type esSearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string                 `json:"_id"`
			Score  float64                `json:"_score"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
