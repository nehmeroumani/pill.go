package auth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func IsAuthenticated(tokenString string) (bool, int, int) {
	if tokenString != "" && tokenString != "deleted" {
		claims := &jwt.StandardClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return GetJWTAuth().publicKey, nil
		})

		if err != nil {
			return false, 0, http.StatusInternalServerError
		} else if !token.Valid {
			return false, 0, http.StatusUnauthorized
		}
		var id int
		if id, err = strconv.Atoi(claims.Subject); err == nil {
			return true, id, http.StatusOK
		} else {
			return false, 0, http.StatusInternalServerError
		}
	} else {
		return false, 0, http.StatusUnauthorized
	}
}

func GetTokenFromRequest(req *http.Request) string {
	// Look for an Authorization header
	if ah := req.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			return ah[7:]
		}
	}

	// Look for "access_token" parameter
	req.ParseMultipartForm(10e6)
	if tokStr := req.Form.Get("access_token"); tokStr != "" {
		return tokStr
	}

	tokenCookie, err := req.Cookie("accessToken")
	if err == nil {
		return tokenCookie.Value
	}
	return ""
}

func SetAccessTokenCookie(w http.ResponseWriter, tokenString string) http.ResponseWriter {
	cookie := http.Cookie{}
	cookie.Name = "accessToken"
	cookie.Value = tokenString
	cookie.HttpOnly = true
	cookie.Path = "/"
	if domainName != "" {
		cookie.Domain = domainName
	}
	if secureToken{
		cookie.Secure = true
	}
	w.Header().Add("Set-Cookie", cookie.String())
	return w
}

func RemoveAccessTokenCookie(w http.ResponseWriter) http.ResponseWriter {
	cookie := http.Cookie{}
	cookie.Name = "accessToken"
	cookie.Value = "deleted"
	cookie.HttpOnly = true
	cookie.Path = "/"
	if domainName != "" {
		cookie.Domain = domainName
	}
	if secureToken {
		cookie.Secure = true
	}
	cookie.Expires, _ = time.Parse("Thu, 01 Jan 1970 00:00:00 GMT", "Thu, 01 Jan 1970 00:00:00 GMT")
	w.Header().Add("Set-Cookie", cookie.String())
	return w
}
