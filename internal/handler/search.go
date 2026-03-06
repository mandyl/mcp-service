package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mandyl/mcp-service/internal/es"
	"github.com/mandyl/mcp-service/internal/model"
)

// SearchHandler handles POST /api/v1/search.
type SearchHandler struct {
	esClient *es.Client
}

// NewSearchHandler creates a SearchHandler.
func NewSearchHandler(esClient *es.Client) *SearchHandler {
	return &SearchHandler{esClient: esClient}
}

// Handle processes the search request.
func (h *SearchHandler) Handle(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    model.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
			Data:    nil,
		})
		return
	}

	if req.Size > 100 {
		c.JSON(http.StatusOK, model.Response{
			Code:    model.CodeBadRequest,
			Message: "size must not exceed 100",
			Data:    nil,
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	data, errCode, err := h.esClient.Search(ctx, &req)
	if err != nil {
		msg := err.Error()
		if errCode == model.CodeIndexNotFound {
			msg = "index not found"
		}
		c.JSON(http.StatusOK, model.Response{
			Code:    errCode,
			Message: msg,
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    model.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// IndicesHandler handles GET /api/v1/indices.
type IndicesHandler struct {
	esClient *es.Client
}

// NewIndicesHandler creates an IndicesHandler.
func NewIndicesHandler(esClient *es.Client) *IndicesHandler {
	return &IndicesHandler{esClient: esClient}
}

// Handle returns the list of available Elasticsearch indices.
func (h *IndicesHandler) Handle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	indices, err := h.esClient.ListIndices(ctx)
	if err != nil {
		c.JSON(http.StatusOK, model.Response{
			Code:    model.CodeESConnectFailed,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    model.CodeSuccess,
		Message: "success",
		Data:    model.IndicesData{Indices: indices},
	})
}

// HealthHandler handles GET /health.
type HealthHandler struct {
	esClient *es.Client
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(esClient *es.Client) *HealthHandler {
	return &HealthHandler{esClient: esClient}
}

// Handle returns the service health status including ES connectivity.
func (h *HealthHandler) Handle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	esOK := h.esClient.Ping(ctx)
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:      "ok",
		ESConnected: esOK,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
	})
}
