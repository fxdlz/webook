package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, l AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	Status   int           `json:"status"`
	RespBody string        `json:"resp_body"`
	Duration time.Duration `json:"duration"`
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, l AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFn: logFn,
	}
}

func (m *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	m.allowReqBody = true
	return m
}

func (m *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	m.allowRespBody = true
	return m
}

func (m *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(c *gin.Context) {
		al := AccessLog{}
		path := c.Request.URL.Path
		if len(path) >= 1024 {
			al.Path = path[:1024]
		}
		al.Method = c.Request.Method
		if m.allowRespBody {
			body, _ := c.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
		}
		start := time.Now()
		if m.allowRespBody {
			c.Writer = &responseWriter{
				al:             &al,
				ResponseWriter: c.Writer,
			}
		}
		defer func() {
			al.Duration = time.Since(start)
			m.logFn(c, al)
		}()
		c.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	if len(data) >= 2048 {
		w.al.RespBody = string(data[:2048])
	} else {
		w.al.RespBody = string(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(code int) {
	w.al.Status = code
	w.ResponseWriter.WriteHeader(code)
}
