package util

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var publicDirPath string
var cacheTTL = 60 * 60 * 24 * 7

func InitXServe(PublicDirPath string, CacheTTL ...int) {
	publicDirPath = filepath.FromSlash(PublicDirPath)
	if CacheTTL != nil && len(CacheTTL) > 0 {
		cacheTTL = CacheTTL[0]
	}
}

func XServe(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var wr CloseableResponseWriter
	requestedFile := r.URL.Path[8:]
	f, err := os.Open(publicDirPath + string(filepath.Separator) + filepath.FromSlash(requestedFile))
	defer f.Close()

	requestedFile = strings.ToLower(requestedFile)

	if err == nil {
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
		} else if strings.HasSuffix(requestedFile, ".gif") {
			contentType = "image/gif"
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
		} else if strings.HasSuffix(requestedFile, ".pdf") {
			contentType = "application/pdf"
		} else if strings.HasSuffix(requestedFile, ".ipa") {
			contentType = "application/octet-stream"
		} else if strings.HasSuffix(requestedFile, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(requestedFile, ".plist") {
			contentType = "application/x-plist"
		} else if strings.HasSuffix(requestedFile, ".apk") {
			contentType = "application/vnd.android.package-archive"
		} else {
			contentType = "text/plain"
		}
		wr = GetResponseWriter(w, r, contentType)
		defer wr.Close()
		wr.SetCacheControl(cacheTTL)
		// if fi, e := f.Stat(); e == nil {
		// 	wr.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
		// }
		cached := false
		if fileInfo, infoErr := f.Stat(); infoErr == nil {
			lastModifiedTime := fileInfo.ModTime().UTC()
			wr.Header().Set("Last-Modified", lastModifiedTime.Format(http.TimeFormat))
			if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
				if modifiedSinceTime, TPErr := time.Parse(http.TimeFormat, modifiedSince); TPErr == nil {
					modifiedSinceTime = modifiedSinceTime.UTC()
					if modifiedSinceTime.Equal(lastModifiedTime) {
						wr.WriteHeader(304)
						wr.Write([]byte{})
						cached = true
					}
				}
			}
		}
		if !cached {
			bufferedReader := bufio.NewReader(f)
			bufferedReader.WriteTo(wr)
		}
	} else {
		w.WriteHeader(404)
		w.Write([]byte(http.StatusText(404)))
	}
}
