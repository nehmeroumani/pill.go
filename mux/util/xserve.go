package util

import (
	"bufio"
	"net/http"
	"os"
	"strings"
)

var publicDirPath string
var cacheTTL int = 3600

func InitXServe(PublicDirPath string, CacheTTL ...int) {
	publicDirPath = PublicDirPath
	if CacheTTL != nil && len(CacheTTL) > 0 {
		cacheTTL = CacheTTL[0]
	}
}

func XServe(w http.ResponseWriter, r *http.Request) {
	var wr CloseableResponseWriter
	requestedFile := r.URL.Path[8:]
	f, err := os.Open(publicDirPath + "/" + requestedFile)
	defer func() {
		f.Close()
		r.Body.Close()
		if wr != nil {
			wr.Close()
		}
	}()
	requestedFile = strings.ToLower(requestedFile)

	if err == nil {
		bufferedReader := bufio.NewReader(f)

		var contentType string
		if strings.HasSuffix(requestedFile, ".js") {
			contentType = "text/javascript"
		} else if strings.HasSuffix(requestedFile, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(requestedFile, ".jpg") {
			contentType = "image/jpg"
		} else if strings.HasSuffix(requestedFile, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(requestedFile, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(requestedFile, ".mp3") {
			contentType = "audio/mpeg"
		} else if strings.HasSuffix(requestedFile, ".ogg") {
			contentType = "audio/ogg"
		} else if strings.HasSuffix(requestedFile, ".woff") {
			contentType = "application/x-font-woff"
		} else if strings.HasSuffix(requestedFile, ".ttf") {
			contentType = "application/font-sfnt"
		} else if strings.HasSuffix(requestedFile, ".svg") {
			contentType = "image/svg+xml"
		} else if strings.HasSuffix(requestedFile, ".eot") {
			contentType = "application/vnd.ms-fontobject"
		} else {
			contentType = "text/plain"
		}
		wr = GetResponseWriter(w, r, contentType)
		wr.SetCacheControl(cacheTTL)
		bufferedReader.WriteTo(wr)
	} else {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
	}
}
