package appy

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type gzipHandler struct {
	config       *Config
	excludedExts map[string]bool
	pool         sync.Pool
}

// Gzip compresses the contents.
func Gzip(config *Config) HandlerFunc {
	return newGzipHandler(config).HandlerFunc
}

func newGzipHandler(config *Config) *gzipHandler {
	handler := &gzipHandler{
		config:       config,
		excludedExts: make(map[string]bool),
		pool: sync.Pool{
			New: func() interface{} {
				gz, _ := gzip.NewWriterLevel(ioutil.Discard, config.HTTPGzipCompressLevel)
				return gz
			},
		},
	}

	for _, e := range config.HTTPGzipExcludedExts {
		handler.excludedExts[e] = true
	}

	return handler
}

func (g *gzipHandler) HandlerFunc(c *Context) {
	if c.Request.Header.Get("Content-Encoding") == "gzip" {
		if c.Request.Body == nil {
			return
		}

		r, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		c.Request.Header.Del("Content-Encoding")
		c.Request.Header.Del("Content-Length")
		c.Request.Body = r
	}

	if g.shouldCompress(c.Request) {
		gz := g.pool.Get().(*gzip.Writer)
		defer g.pool.Put(gz)
		defer gz.Reset(ioutil.Discard)
		gz.Reset(c.Writer)

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		c.Writer = &gzipWriter{c.Writer, gz}
		defer func() {
			gz.Close()
			c.Header("Content-Length", fmt.Sprint(c.Writer.Size()))
		}()
	}

	c.Next()
}

func (g *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Content-Type"), "text/event-stream") {
		return false
	}

	ext := filepath.Ext(req.URL.Path)
	if _, ok := g.excludedExts[ext]; ok {
		return false
	}

	for _, p := range g.config.HTTPGzipExcludedPaths {
		if strings.Contains(req.URL.Path, p) {
			return false
		}
	}

	return true
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}
