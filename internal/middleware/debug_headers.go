package middleware

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const debugHeaderEnv = "MCP_DEBUG"

// DebugHeaders injects diagnostic response headers when MCP_DEBUG=true.
//
// Headers added:
//   X-Request-ID  — unique per-request identifier (Unix nanosecond timestamp)
//   X-Handler     — the matched route handler path
//   X-Latency-Ms  — request processing time in milliseconds
//   X-Debug       — literal "true" to indicate debug mode is active
func DebugHeaders() gin.HandlerFunc {
	enabled := os.Getenv(debugHeaderEnv) == "true"
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		start := time.Now()
		requestID := start.UnixNano()

		c.Next()

		latency := time.Since(start).Milliseconds()
		c.Header("X-Debug", "true")
		c.Header("X-Request-ID", itoa(requestID))
		c.Header("X-Handler", c.FullPath())
		c.Header("X-Latency-Ms", itoa(int64(latency)))
	}
}

// itoa converts an int64 to a decimal string without importing strconv
// (keeps the import list minimal).
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
