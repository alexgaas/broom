package api

import (
	"encoding/json"
	"fmt"
	"sync"

	"cache/internal/bch"
	"cache/internal/config"
	"cache/internal/log"
	"cache/internal/metrics"

	"github.com/gin-gonic/gin"
)

type Api struct {
	Opts *config.ConfYaml
	Log  *log.TLog

	Core       CoreInterface
	BloomCache *bch.BloomCache
	Metrics    *metrics.Metrics
}

// CoreInterface defines the interface for core dependency
type CoreInterface interface{}

func CreateApi(opts *config.ConfYaml, logger *log.TLog) (*Api, error) {
	var api Api
	api.Opts = opts
	api.Log = logger
	return &api, nil
}

func (a *Api) RunHTTPServer() error {
	id := "(http) (server)"

	r := a.routerEngine()

	port := a.Opts.Server.Port
	if port > 0 {
		a.Log.Info(fmt.Sprintf("%s run http server on port:%d", id, port))
		return r.Run(fmt.Sprintf(":%d", port))
	}

	return nil
}

func (a *Api) routerEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(a.apiRequestLogger)

	if a.Metrics != nil {
		r.Use(a.Metrics.Middleware())
		r.GET("/metrics", a.Metrics.Handler())
	}

	r.GET("/health", a.getHealth)

	// Bloom filter cache endpoints
	bloom := r.Group("/bloom")
	{
		bloom.POST("/add", a.bloomAdd)
		bloom.GET("/check/:key", a.bloomCheck)
		bloom.POST("/check", a.bloomCheckPost)
		bloom.GET("/stats", a.bloomStats)
		bloom.POST("/clear", a.bloomClear)
	}

	return r
}

func (a *Api) Apiloop(wg *sync.WaitGroup) {
	defer wg.Done()
	a.RunHTTPServer()
}

func (a *Api) apiRequestLogger(c *gin.Context) {
	path, ip := a.apiRequestString(c)
	a.Log.Info(fmt.Sprintf("%s '%s %s' %d %s", ip, c.Request.Method, path,
		c.Writer.Status(), c.Request.UserAgent()))
	c.Next()
}

func (a *Api) apiRequestString(c *gin.Context) (string, string) {
	ip := c.ClientIP()
	if len(ip) == 0 {
		ip = "socket"
	}

	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}

	return path, ip
}

type ResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Success  bool           `json:"success"`
	Errors   []ResponseInfo `json:"errors,omitempty"`
	Messages []ResponseInfo `json:"messages,omitempty"`
}

func (a *Api) apiSendError(c *gin.Context, code int, errStr string) {
	var r Response

	r.Success = false
	var ri ResponseInfo
	ri.Code = code
	ri.Message = errStr
	r.Errors = append(r.Errors, ri)

	response, _ := json.Marshal(r)

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(code, string(response))
}

func (a *Api) apiSendOK(c *gin.Context, code int, msgStr string) {
	var r Response

	r.Success = true
	var ri ResponseInfo
	ri.Code = code
	ri.Message = msgStr
	r.Messages = append(r.Messages, ri)

	responseBody, _ := json.Marshal(r)

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(code, string(responseBody))
}
