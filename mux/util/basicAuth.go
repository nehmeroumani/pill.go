package util

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func GetBasicAuthData(w http.ResponseWriter, r *http.Request, customHeader ...string) (string, string) {
	authorization := ""
	if len(r.Header) > 0 {
		header := "Authorization"
		if customHeader != nil && len(customHeader) > 0 {
			header = customHeader[0]
		}
		if r.Header[header] != nil && len(r.Header[header]) > 0 {
			authorization = r.Header[header][0]
		} else {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return "", ""
		}
	} else {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
		return "", ""
	}
	auth := strings.SplitN(authorization, " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		http.Error(w, "bad syntax", http.StatusBadRequest)
		return "", ""
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) > 1 {
		return pair[0], pair[1]
	}
	return "", ""
}
