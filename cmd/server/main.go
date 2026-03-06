package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mandyl/mcp-service/internal/es"
	"github.com/mandyl/mcp-service/internal/handler"
	"github.com/mandyl/mcp-service/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	esClient, err := es.New(cfg.ESAddr)
	if err != nil {
		log.Fatalf("Failed to create ES client: %v", err)
	}
	log.Printf("ES address: %s", cfg.ESAddr)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Health check (no auth required, called by K8s probe).
	r.GET("/health", handler.NewHealthHandler(esClient).Handle)

	// API v1 routes.
	v1 := r.Group("/api/v1")
	{
		v1.POST("/search", handler.NewSearchHandler(esClient).Handle)
		v1.GET("/indices", handler.NewIndicesHandler(esClient).Handle)
	}

	addr := ":" + cfg.Port
	log.Printf("mcp-service listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
