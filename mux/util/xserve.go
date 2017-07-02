package util

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var publicLocalPath, publicCloudUrlPath, appVersion string
var filesFromCloud bool
var cacheTTL = 60 * 60 * 24 * 7

func InitXServe(PublicLocalPath string, opts ...interface{}) {
	publicLocalPath = filepath.FromSlash(PublicLocalPath)
	filesFromCloud = false
	if opts != nil && len(opts) == 1 {
		cacheTTL = opts[0].(int)
	} else {
		if opts != nil {
			if len(opts) > 0 {
				publicCloudUrlPath = opts[0].(string)
			}
			if len(opts) > 1 {
				filesFromCloud = opts[1].(bool)
			}
			if len(opts) > 2 {
				appVersion = opts[2].(string)
			}
		}
	}
}

func XServe(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	requestedFile := r.URL.Path[8:]
	if filesFromCloud && publicCloudUrlPath != "" {
		d := "http://www.example.com/"
		u, err := url.Parse(d + requestedFile)
		if err == nil {
			if u.RawQuery != "" {
				u.RawQuery += "&"
			}
			u.RawQuery += "app_version=" + url.QueryEscape(appVersion)
			requestedFile = strings.Replace(u.String(), d, "", -1)
		}
		fmt.Println(requestedFile)
		http.Redirect(w, r, publicCloudUrlPath+"/"+requestedFile, 307)
	} else {
		f, err := os.Open(publicLocalPath + string(filepath.Separator) + filepath.FromSlash(requestedFile))
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
			w.Header().Set("Content-Type", contentType)
			setResponseWriterCacheControl(w, cacheTTL)
			cached := false
			if fileInfo, infoErr := f.Stat(); infoErr == nil {
				w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
				lastModifiedTime := fileInfo.ModTime().UTC()
				w.Header().Set("Last-Modified", lastModifiedTime.Format(http.TimeFormat))
				if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
					if modifiedSinceTime, TPErr := time.Parse(http.TimeFormat, modifiedSince); TPErr == nil {
						modifiedSinceTime = modifiedSinceTime.UTC()
						if modifiedSinceTime.Equal(lastModifiedTime) {
							w.WriteHeader(304)
							w.Write([]byte{})
							cached = true
						}
					}
				}
			}
			if !cached {
				bufferedReader := bufio.NewReader(f)
				bufferedReader.WriteTo(w)
			}
		} else {
			w.WriteHeader(404)
			w.Write([]byte(http.StatusText(404)))
		}
	}
}
