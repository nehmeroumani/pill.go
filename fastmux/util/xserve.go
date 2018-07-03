package util

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nehmeroumani/pill.go/helpers"
	"github.com/valyala/fasthttp"
)

var (
	staticFilesPath, uploadsPath, appVersion string
	staticFilesURLPath, uploadsURLPath       string
	staticFilesFromCloud, uploadsFromCloud   bool
	cacheTTL                                 = time.Second * 60 * 60 * 24 * 30 * 6
	fsUploadsHandler, fsStaticFilesHandler   fasthttp.RequestHandler
)

func InitXServe(StaticFilesPath string, StaticFilesURLPath string, StaticFilesFromCloud bool, UploadsPath string, UploadsURLPath string, UploadsFromCloud bool, AppVersion string, opts ...bool) {
	staticFilesPath = StaticFilesPath
	staticFilesURLPath = StaticFilesURLPath
	staticFilesFromCloud = StaticFilesFromCloud
	uploadsPath = UploadsPath
	uploadsURLPath = UploadsURLPath
	uploadsFromCloud = UploadsFromCloud
	appVersion = AppVersion
	generateIndexPages := false
	if opts != nil && len(opts) > 0 {
		generateIndexPages = opts[0]
	}
	fsUploads := &fasthttp.FS{
		Root:                 "/",
		GenerateIndexPages:   generateIndexPages,
		IndexNames:           []string{"index.html"},
		Compress:             true,
		AcceptByteRange:      true,
		CompressedFileSuffix: ".gz",
		PathRewrite: fasthttp.PathRewriteFunc(func(requestCtx *fasthttp.RequestCtx) []byte {
			requestedFile := helpers.BytesToString(requestCtx.Path())
			if strings.HasPrefix(requestedFile, "/public") {
				requestedFile = requestedFile[len("/public"):]
			}
			requestedFile = requestedFile[len("/uploads"):]
			return []byte(path.Join(uploadsPath, requestedFile))
		}),
	}
	fsUploadsHandler = fsUploads.NewRequestHandler()
	fsStaticFiles := &fasthttp.FS{
		Root:                 "/",
		IndexNames:           []string{"index.html"},
		GenerateIndexPages:   generateIndexPages,
		Compress:             true,
		AcceptByteRange:      true,
		CompressedFileSuffix: ".gz",
		PathRewrite: fasthttp.PathRewriteFunc(func(requestCtx *fasthttp.RequestCtx) []byte {
			requestedFile := helpers.BytesToString(requestCtx.Path())
			if strings.HasPrefix(requestedFile, "/public") {
				requestedFile = requestedFile[len("/public"):]
			}
			requestedFile = requestedFile[len("/static"):]
			return []byte(path.Join(staticFilesPath, requestedFile))
		}),
	}
	fsStaticFilesHandler = fsStaticFiles.NewRequestHandler()

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
	} else if f, err := os.Open(filesPath + filepath.FromSlash(requestedFile)); err == nil {
		defer f.Close()
		fileInfo, infoErr := f.Stat()
		if infoErr != nil {
			requestCtx.Redirect("/404", 307)
		} else {
			requestCtx.Response.Header.Set("Cache-Control", "public, max-age="+strconv.FormatInt(int64(cacheTTL), 10))
			if requestCtx.IsHead() && !fileInfo.IsDir() {
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
					case ".mp4":
						contentType = "video/mp4"
					case ".avi":
						contentType = "video/x-msvideo"
					case ".wmv":
						contentType = "video/x-ms-wmv"
					case ".3gp":
						contentType = "video/3gpp"
					case ".mov":
						contentType = "video/quicktime"
					case ".mkv":
						contentType = "video/x-matroska"
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
					fileSize := fileInfo.Size()
					requestCtx.Response.Header.Set("Content-Length", strconv.FormatInt(fileSize, 10))
					requestCtx.Response.Header.Set("Accept-Ranges", "bytes")
					requestCtx.SetStatusCode(302)
					requestCtx.Write([]byte{})
				}
			} else {
				requestCtx.Request.SetRequestURIBytes(requestCtx.Path())
				if isStatic {
					fsStaticFilesHandler(requestCtx)
				} else {
					fsUploadsHandler(requestCtx)
				}
			}
		}
	} else {
		requestCtx.Redirect("/404", 307)
	}
}
