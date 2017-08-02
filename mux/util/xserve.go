package util

import (
	"bufio"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var staticFilesPath, uploadsPath, appVersion string
var staticFilesURLPath, uploadsURLPath string
var staticFilesFromCloud, uploadsFromCloud bool
var cacheTTL = 60 * 60 * 24 * 7

func InitXServe(StaticFilesPath string, StaticFilesURLPath string, StaticFilesFromCloud bool, UploadsPath string, UploadsURLPath string, UploadsFromCloud bool, AppVersion string) {
	staticFilesPath = StaticFilesPath
	staticFilesURLPath = StaticFilesURLPath
	staticFilesFromCloud = StaticFilesFromCloud
	uploadsPath = UploadsPath
	uploadsURLPath = UploadsURLPath
	uploadsFromCloud = UploadsFromCloud
	appVersion = AppVersion
}

func StaticFilesServe(w http.ResponseWriter, r *http.Request) {
	xServe(w, r, staticFilesPath, staticFilesURLPath, staticFilesFromCloud, true)
}

func UploadsServe(w http.ResponseWriter, r *http.Request) {
	xServe(w, r, uploadsPath, uploadsURLPath, uploadsFromCloud, false)
}
func xServe(w http.ResponseWriter, r *http.Request, filesPath string, filesURLPath string, fromCloud bool, isStatic bool) {
	defer r.Body.Close()
	requestedFile := r.URL.Path
	if strings.HasPrefix(requestedFile, "/public") {
		requestedFile = requestedFile[len("/public"):]
	}
	if isStatic {
		requestedFile = requestedFile[len("/static"):]
	} else {
		requestedFile = requestedFile[len("/uploads"):]
	}
	if fromCloud {
		requestedFile += "?" + r.URL.RawQuery
		if r.URL.RawQuery != "" {
			requestedFile += "&"
		}
		requestedFile += "app_version=" + url.QueryEscape(appVersion)
		http.Redirect(w, r, filesURLPath+requestedFile, 307)
	} else {
		f, err := os.Open(filesPath + filepath.FromSlash(requestedFile))
		defer f.Close()

		fileExtension := strings.ToLower(filepath.Ext(requestedFile))

		if err == nil {
			var contentType string
			switch fileExtension {
			case ".js":
				contentType = "text/javascript"
			case ".css":
				contentType = "text/css"
			case ".jpg":
				contentType = "image/jpg"
			case ".png":
				contentType = "image/png"
			case ".jpeg":
				contentType = "image/jpeg"
			case ".gif":
				contentType = "image/gif"
			case ".mp3":
				contentType = "audio/mpeg"
			case ".ogg":
				contentType = "audio/ogg"
			case ".woff":
				contentType = "application/x-font-woff"
			case ".ttf":
				contentType = "application/font-sfnt"
			case ".svg":
				contentType = "image/svg+xml"
			case ".eot":
				contentType = "application/vnd.ms-fontobject"
			case ".pdf":
				contentType = "application/pdf"
			case ".ipa":
				contentType = "application/octet-stream"
			case ".html":
				contentType = "text/html"
			case ".plist":
				contentType = "application/x-plist"
			case ".apk":
				contentType = "application/vnd.android.package-archive"
			default:
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
