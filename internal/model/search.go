package model

// SearchRequest is the body for POST /api/v1/search.
type SearchRequest struct {
	Index   string         `json:"index" binding:"required"`
	Query   string         `json:"query" binding:"required"`
	From    int            `json:"from"`
	Size    int            `json:"size"`
	Filters *FilterOptions `json:"filters,omitempty"`
}

// FilterOptions provides optional field-level filtering.
type FilterOptions struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// Hit represents a single Elasticsearch document hit.
type Hit struct {
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

// SearchData is the payload inside a successful search response.
type SearchData struct {
	Total int   `json:"total"`
	Hits  []Hit `json:"hits"`
	From  int   `json:"from"`
	Size  int   `json:"size"`
}

// Response is the standard API response envelope.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// IndicesData is the payload for the indices response.
type IndicesData struct {
	Indices []string `json:"indices"`
}

// HealthResponse is the payload for /health.
type HealthResponse struct {
	Status      string `json:"status"`
	ESConnected bool   `json:"es_connected"`
	Timestamp   string `json:"timestamp"`
}

// Error codes per PRD §4.2.3.
const (
	CodeSuccess           = 0
	CodeIndexNotFound     = 40001
	CodeBadRequest        = 40002
	CodeQuerySyntaxError  = 40003
	CodeESConnectFailed   = 50001
	CodeInternalError     = 50002
)
