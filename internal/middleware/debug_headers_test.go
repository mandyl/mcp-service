package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DebugHeaders())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestDebugHeaders_Disabled(t *testing.T) {
	t.Setenv("MCP_DEBUG", "")
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	if v := w.Header().Get("X-Debug"); v != "" {
		t.Errorf("expected no X-Debug header when disabled, got %q", v)
	}
}

func TestDebugHeaders_Enabled(t *testing.T) {
	t.Setenv("MCP_DEBUG", "true")
	// Re-init so the middleware picks up the new env value.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(DebugHeaders())
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	for _, hdr := range []string{"X-Debug", "X-Request-ID", "X-Handler", "X-Latency-Ms"} {
		if w.Header().Get(hdr) == "" {
			t.Errorf("expected header %s to be set when MCP_DEBUG=true", hdr)
		}
	}
	if got := w.Header().Get("X-Debug"); got != "true" {
		t.Errorf("X-Debug: want %q, got %q", "true", got)
	}
	if got := w.Header().Get("X-Handler"); got != "/health" {
		t.Errorf("X-Handler: want %q, got %q", "/health", got)
	}
}
