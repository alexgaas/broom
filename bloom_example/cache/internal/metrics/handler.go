package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/common/expfmt"
)

func (m *Metrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		mfs, err := m.Registry.Gather()
		if err != nil {
			c.String(http.StatusInternalServerError, "error gathering metrics: %v", err)
			return
		}

		contentType := expfmt.Negotiate(c.Request.Header)
		c.Header("Content-Type", string(contentType))
		c.Status(http.StatusOK)

		enc := expfmt.NewEncoder(c.Writer, contentType)
		for _, mf := range mfs {
			if err := enc.Encode(mf); err != nil {
				return
			}
		}

		if closer, ok := enc.(expfmt.Closer); ok {
			closer.Close()
		}
	}
}
