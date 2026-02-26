package metrics

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a gin.HandlerFunc that records HTTP request metrics:
// request count, duration, and response status class.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		statusStr := fmt.Sprintf("%d", status)
		duration := time.Since(start)

		m.HTTPRequestsTotal.With(map[string]string{
			"method": method,
			"path":   path,
			"status": statusStr,
		}).Inc()

		m.HTTPRequestDuration.With(map[string]string{
			"method": method,
			"path":   path,
		}).RecordDuration(duration)

		statusClass := fmt.Sprintf("%dxx", status/100)
		m.HTTPResponseStatusTotal.With(map[string]string{
			"status_class": statusClass,
		}).Inc()
	}
}
