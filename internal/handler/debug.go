package handler

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

// DebugHeadersHandler handles GET /api/v1/debug/headers.
// It returns all HTTP request headers received by the MCP service,
// which is useful for verifying that Higress correctly injects headers
// such as x-user-id after ext-auth validation.
//
// ⚠️  Security note: this endpoint exposes all request headers, including
// sensitive auth tokens. It is intended for debugging only and should be
// disabled or restricted by IP in production environments.
type DebugHeadersHandler struct{}

// NewDebugHeadersHandler creates a DebugHeadersHandler.
func NewDebugHeadersHandler() *DebugHeadersHandler {
	return &DebugHeadersHandler{}
}

// Handle returns all request headers as a JSON map.
func (h *DebugHeadersHandler) Handle(c *gin.Context) {
	headers := make(map[string]string, len(c.Request.Header))
	for key := range c.Request.Header {
		// Use canonical form; take only the first value for simplicity.
		headers[http.CanonicalHeaderKey(key)] = c.Request.Header.Get(key)
	}

	// Sort keys for deterministic output (easier to read in tests).
	sorted := make(map[string]string, len(headers))
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sorted[k] = headers[k]
	}

	c.JSON(http.StatusOK, gin.H{
		"headers": sorted,
	})
}
