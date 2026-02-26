package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// health check
func (a *Api) getHealth(c *gin.Context) {
	if a.Core == nil {
		a.apiSendError(c, 502, "Internal error")
		return
	}

	a.apiSendOK(c, 200, "ok")
}

// Bloom filter handlers

type BloomAddRequest struct {
	Key  string   `json:"key,omitempty"`
	Keys []string `json:"keys,omitempty"`
}

type BloomCheckRequest struct {
	Key string `json:"key"`
}

type BloomCheckResponse struct {
	Key    string `json:"key"`
	Exists bool   `json:"exists"`
}

// bloomAdd adds key(s) to the bloom filter
func (a *Api) bloomAdd(c *gin.Context) {
	if a.BloomCache == nil {
		a.apiSendError(c, http.StatusServiceUnavailable, "Bloom cache not initialized")
		return
	}

	var req BloomAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.apiSendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	start := time.Now()

	count := 0
	if req.Key != "" {
		a.BloomCache.Add(req.Key)
		count++
	}

	for _, key := range req.Keys {
		if key != "" {
			a.BloomCache.Add(key)
			count++
		}
	}

	if count == 0 {
		a.apiSendError(c, http.StatusBadRequest, "No keys provided")
		return
	}

	if a.Metrics != nil {
		a.Metrics.BloomAddsTotal.Inc()
		a.Metrics.BloomKeysAddedTotal.Add(int64(count))
		a.Metrics.BloomOperationDuration.With(map[string]string{"operation": "add"}).RecordDuration(time.Since(start))
	}

	a.apiSendOK(c, http.StatusOK, "added")
}

// bloomCheck checks if a key exists (GET with URL param)
func (a *Api) bloomCheck(c *gin.Context) {
	if a.BloomCache == nil {
		a.apiSendError(c, http.StatusServiceUnavailable, "Bloom cache not initialized")
		return
	}

	key := c.Param("key")
	if key == "" {
		a.apiSendError(c, http.StatusBadRequest, "Key is required")
		return
	}

	start := time.Now()
	exists := a.BloomCache.Has(key)

	if a.Metrics != nil {
		a.Metrics.BloomChecksTotal.Inc()
		if exists {
			a.Metrics.BloomHitsTotal.Inc()
		} else {
			a.Metrics.BloomMissesTotal.Inc()
		}
		a.Metrics.BloomOperationDuration.With(map[string]string{"operation": "check"}).RecordDuration(time.Since(start))
	}

	c.JSON(http.StatusOK, BloomCheckResponse{
		Key:    key,
		Exists: exists,
	})
}

// bloomCheckPost checks if a key exists (POST with JSON body)
func (a *Api) bloomCheckPost(c *gin.Context) {
	if a.BloomCache == nil {
		a.apiSendError(c, http.StatusServiceUnavailable, "Bloom cache not initialized")
		return
	}

	var req BloomCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.apiSendError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		a.apiSendError(c, http.StatusBadRequest, "Key is required")
		return
	}

	start := time.Now()
	exists := a.BloomCache.Has(req.Key)

	if a.Metrics != nil {
		a.Metrics.BloomChecksTotal.Inc()
		if exists {
			a.Metrics.BloomHitsTotal.Inc()
		} else {
			a.Metrics.BloomMissesTotal.Inc()
		}
		a.Metrics.BloomOperationDuration.With(map[string]string{"operation": "check"}).RecordDuration(time.Since(start))
	}

	c.JSON(http.StatusOK, BloomCheckResponse{
		Key:    req.Key,
		Exists: exists,
	})
}

// bloomStats returns bloom filter statistics
func (a *Api) bloomStats(c *gin.Context) {
	if a.BloomCache == nil {
		a.apiSendError(c, http.StatusServiceUnavailable, "Bloom cache not initialized")
		return
	}

	stats := a.BloomCache.Stats()
	c.JSON(http.StatusOK, stats)
}

// bloomClear clears the bloom filter
func (a *Api) bloomClear(c *gin.Context) {
	if a.BloomCache == nil {
		a.apiSendError(c, http.StatusServiceUnavailable, "Bloom cache not initialized")
		return
	}

	a.BloomCache.Clear()

	if a.Metrics != nil {
		a.Metrics.BloomClearsTotal.Inc()
	}

	a.apiSendOK(c, http.StatusOK, "cleared")
}
