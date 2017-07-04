package util

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/valyala/fasthttp"
)

func GetBasicAuthData(requestCtx *fasthttp.RequestCtx, customHeader ...string) (string, string) {
	authorization := ""
	if len(requestCtx.Request.Header) > 0 {
		header := "Authorization"
		if customHeader != nil && len(customHeader) > 0 {
			header = customHeader[0]
		}
		if requestCtx.Request.Header[header] != nil && len(requestCtx.Request.Header[header]) > 0 {
			authorization = requestCtx.Request.Header[header][0]
		} else {
			requestCtx.Error("authorization failed", http.StatusUnauthorized)
			return "", ""
		}
	} else {
		requestCtx.Error("authorization failed", http.StatusUnauthorized)
		return "", ""
	}
	auth := strings.SplitN(authorization, " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		requestCtx.Error("bad syntax", http.StatusBadRequest)
		return "", ""
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) > 1 {
		return pair[0], pair[1]
	}
	return "", ""
}
