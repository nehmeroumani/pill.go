package util

import (
	"bufio"
	"bytes"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nehmeroumani/pill.go/helpers"
	"github.com/valyala/fasthttp"
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

func StaticFilesServe(requestCtx *fasthttp.RequestCtx) {
	xServe(requestCtx, staticFilesPath, staticFilesURLPath, staticFilesFromCloud, true)
}

func UploadsServe(requestCtx *fasthttp.RequestCtx) {
	xServe(requestCtx, uploadsPath, uploadsURLPath, uploadsFromCloud, false)
}
func xServe(requestCtx *fasthttp.RequestCtx, filesPath string, filesURLPath string, fromCloud bool, isStatic bool) {
	requestedFile := helpers.BytesToString(requestCtx.Path())
	if strings.HasPrefix(requestedFile, "/public") {
		requestedFile = requestedFile[len("/public"):]
	}
	if isStatic {
		requestedFile = requestedFile[len("/static"):]
	} else {
		requestedFile = requestedFile[len("/uploads"):]
	}
	if fromCloud {
		queryString := helpers.BytesToString(requestCtx.URI().QueryString())
		requestedFile += "?" + queryString
		if queryString != "" {
			requestedFile += "&"
		}
		requestedFile += "app_version=" + url.QueryEscape(appVersion)
		requestCtx.Redirect(filesURLPath+requestedFile, 307)
	} else {
		f, err := os.Open(filesPath + filepath.FromSlash(requestedFile))
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
			requestCtx.SetContentType(contentType)
			setResponseWriterCacheControl(requestCtx, cacheTTL)
			cached := false
			if fileInfo, infoErr := f.Stat(); infoErr == nil {
				requestCtx.Response.Header.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
				lastModifiedTime := fileInfo.ModTime().UTC()
				requestCtx.Response.Header.Set("Last-Modified", lastModifiedTime.Format(http.TimeFormat))
				if modifiedSince := helpers.BytesToString(requestCtx.Request.Header.Peek("If-Modified-Since")); modifiedSince != "" {
					if modifiedSinceTime, TPErr := time.Parse(http.TimeFormat, modifiedSince); TPErr == nil {
						modifiedSinceTime = modifiedSinceTime.UTC()
						if modifiedSinceTime.Equal(lastModifiedTime) {
							requestCtx.SetStatusCode(304)
							requestCtx.Response.SetBodyStream(bytes.NewReader([]byte{}), 0)
							cached = true
						}
					}
				}
				if !cached {
					buf := bufio.NewReader(f)
					requestCtx.Response.SetBodyStream(buf, int(fileInfo.Size()))
				}
			}
		} else {
			requestCtx.SetStatusCode(404)
			requestCtx.Response.SetBodyStream(bytes.NewReader([]byte(http.StatusText(404))), len([]byte(http.StatusText(404))))
		}
	}
}
func setResponseWriterCacheControl(requestCtx *fasthttp.RequestCtx, TLL int, opts ...bool) {
	if TLL > 0 {
		private := false
		if opts != nil && len(opts) > 0 {
			private = opts[0]
		}
		cacheControl := ""
		if private {
			cacheControl += "private, "
		}
		cacheControl += "max-age=" + strconv.Itoa(TLL)
		requestCtx.Response.Header.Set("Cache-Control", cacheControl)
	}
}
