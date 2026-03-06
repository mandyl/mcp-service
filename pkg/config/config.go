package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the service configuration.
type Config struct {
	Port   string
	ESAddr string // e.g. "http://elasticsearch.backend.svc.cluster.local:9200"
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	host := getEnv("ES_HOST", "elasticsearch.backend.svc.cluster.local")
	portStr := getEnv("ES_PORT", "9200")
	if _, err := strconv.Atoi(portStr); err != nil {
		return nil, fmt.Errorf("invalid ES_PORT %q: %w", portStr, err)
	}

	esAddr := getEnv("ES_ADDR", fmt.Sprintf("http://%s:%s", host, portStr))

	return &Config{
		Port:   getEnv("PORT", "8080"),
		ESAddr: esAddr,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
