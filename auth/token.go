package auth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func IsAuthenticated(tokenString string, opts ...int64) (bool, int, int) {
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
		var lastPasswordUpdate int64
		if opts != nil && len(opts) > 0 {
			lastPasswordUpdate = opts[0]
		}
		if claims.IssuedAt > lastPasswordUpdate {
			if getTokenRemainingValidity(claims.ExpiresAt) > 0 {
				var id int
				if id, err = strconv.Atoi(claims.Subject); err == nil {
					return true, id, http.StatusOK
				}
				return false, 0, http.StatusInternalServerError
			}
		}
	}
	return false, 0, http.StatusUnauthorized
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

	tokenCookie, err := req.Cookie("access_token")
	if err == nil {
		return tokenCookie.Value
	}
	return ""
}

func SetAccessTokenCookie(w http.ResponseWriter, tokenString string, opts ...bool) http.ResponseWriter {
	cookie := http.Cookie{}
	cookie.Name = "access_token"
	cookie.Value = tokenString
	cookie.HttpOnly = true
	cookie.Path = "/"
	if domainName != "" {
		cookie.Domain = domainName
	}
	if secureToken {
		cookie.Secure = true
	}
	if opts != nil && len(opts) > 0 {
		if opts[0] {
			cookie.Expires = time.Now().Add(time.Hour * tokenDuration)
		}
	}
	w.Header().Add("Set-Cookie", cookie.String())
	return w
}

func RefreshAccessTokenCookie(w http.ResponseWriter, req *http.Request, userID int) http.ResponseWriter {
	tokenWasRefurbished, _ := req.Cookie("token_was_refurbished")
	refresh := true
	if tokenWasRefurbished != nil {
		if r, parseErr := strconv.ParseBool(tokenWasRefurbished.Value); parseErr == nil {
			refresh = !r
		}
	}
	if refresh {
		cookie, err := req.Cookie("access_token")
		if err == nil {
			jwtAuth := GetJWTAuth()
			tokenString, tErr := jwtAuth.GenerateToken(userID)
			if tErr == nil {
				tokenWasRefurbishedCookie := http.Cookie{}
				tokenWasRefurbishedCookie.Name = "token_was_refurbished"
				tokenWasRefurbishedCookie.Value = "true"
				tokenWasRefurbishedCookie.HttpOnly = true
				tokenWasRefurbishedCookie.Path = "/"
				cookie.Value = tokenString
				cookie.HttpOnly = true
				cookie.Path = "/"
				if domainName != "" {
					cookie.Domain = domainName
					tokenWasRefurbishedCookie.Domain = domainName
				}
				if secureToken {
					tokenWasRefurbishedCookie.Secure = true
					cookie.Secure = true
				}
				if cookie.MaxAge > 0 {
					cookie.Expires = time.Now().Add(time.Hour * tokenDuration)
				}
				w.Header().Add("Set-Cookie", cookie.String())
				w.Header().Add("Set-Cookie", tokenWasRefurbishedCookie.String())
			}
		}
	}
	return w
}

func RemoveAccessTokenCookie(w http.ResponseWriter) http.ResponseWriter {
	cookie := http.Cookie{}
	cookie.Name = "access_token"
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

func getTokenRemainingValidity(timestamp int64) int {
	tm := time.Unix(timestamp, 0)
	remainder := tm.Sub(time.Now())
	if remainder > 0 {
		return int(remainder.Seconds())
	}
	return -1
}
