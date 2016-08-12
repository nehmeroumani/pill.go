package util

import (
	"compress/gzip"
	"net/http"
	"strconv"
	"strings"
)

type CloseableResponseWriter interface {
	http.ResponseWriter
	Close()
	SetContentType(string)
	SetCacheControl(int)
	SetStatusCode(int)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	*gzip.Writer
}

func (this gzipResponseWriter) Write(data []byte) (int, error) {
	return this.Writer.Write(data)
}

func (this gzipResponseWriter) Close() {
	this.Writer.Close()
}

func (this gzipResponseWriter) SetContentType(contentType string) {
	this.ResponseWriter.Header().Set("Content-Type", contentType)
}

func (this gzipResponseWriter) SetCacheControl(TLL int) {
	if TLL > 0 {
		this.ResponseWriter.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(TLL))
	}
}

func (this gzipResponseWriter) SetStatusCode(statusCode int) {
	this.ResponseWriter.WriteHeader(statusCode)
}

func (this gzipResponseWriter) Header() http.Header {
	return this.ResponseWriter.Header()
}

type closeableResponseWriter struct {
	http.ResponseWriter
}

func (this closeableResponseWriter) Close() {
}

func (this closeableResponseWriter) SetContentType(contentType string) {
	this.ResponseWriter.Header().Set("Content-Type", contentType)
}

func (this closeableResponseWriter) SetCacheControl(TLL int) {
	if TLL > 0 {
		this.ResponseWriter.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(TLL))
	}
}

func (this closeableResponseWriter) SetStatusCode(statusCode int) {
	this.ResponseWriter.WriteHeader(statusCode)
}

func GetResponseWriter(w http.ResponseWriter, r *http.Request, contentType string) CloseableResponseWriter {
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		return gzipResponseWriter{ResponseWriter: w,
			Writer: gzip.NewWriter(w)}
	} else {
		return closeableResponseWriter{ResponseWriter: w}
	}
}
