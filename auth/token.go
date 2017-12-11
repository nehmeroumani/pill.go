package auth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nehmeroumani/pill.go/helpers"
	"github.com/valyala/fasthttp"
)

func IsAuthenticated(tokenString string, opts ...int64) (bool, int, string) {
	if tokenString != "" && tokenString != "deleted" {
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return GetJWTAuth().publicKey, nil
		})

		if err != nil || !token.Valid {
			return false, 0, ""
		}
		var lastPasswordUpdate int64
		if opts != nil && len(opts) > 0 {
			lastPasswordUpdate = opts[0]
		}
		if claims.IssuedAt > lastPasswordUpdate {
			if getTokenRemainingValidity(claims.ExpiresAt) > 0 {
				var id int
				if id, err = strconv.Atoi(claims.Subject); err == nil {
					return true, id, claims.Role
				}
			}
		}
	}
	return false, 0, ""
}

func GetTokenFromRequest(w http.ResponseWriter, req *http.Request) string {
	// Look for an Authorization header
	if ah := req.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			return ah[7:]
		}
	}
	tokenCookie, err := req.Cookie("access_token")
	if err == nil {
		if tokenCookie.Value != "" {
			return tokenCookie.Value
		}
	}
	// Look for "access_token" parameter
	req.ParseMultipartForm(10e6)
	if tokStr := req.Form.Get("access_token"); tokStr != "" {
		SetAccessTokenCookie(w, tokStr)
		return tokStr
	}
	return ""
}

func GetTokenFromFastHttpRequest(requestCtx *fasthttp.RequestCtx) string {
	// Look for an Authorization header
	if ah := helpers.BytesToString(requestCtx.Request.Header.Peek("Authorization")); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:7]) == "BEARER " {
			return ah[7:]
		}
	}
	tokenCookie := requestCtx.Request.Header.Cookie("access_token")
	if tokenCookie != nil {
		return string(tokenCookie)
	}
	// Look for "access_token" parameter
	if tokStr := helpers.BytesToString(requestCtx.QueryArgs().Peek("access_token")); tokStr != "" {
		SetAccessTokenFastHttpCookie(requestCtx, tokStr)
		return tokStr
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
			rememberMeCookie := http.Cookie{}
			rememberMeCookie.Value = "true"
			rememberMeCookie.Name = "remember_me"
			rememberMeCookie.HttpOnly = true
			rememberMeCookie.Path = "/"
			rememberMeCookie.Expires = time.Now().Add(time.Hour * tokenDuration)
			if domainName != "" {
				rememberMeCookie.Domain = domainName
			}
			if secureToken {
				rememberMeCookie.Secure = true
			}
			w.Header().Add("Set-Cookie", rememberMeCookie.String())
		}
	}
	w.Header().Add("Set-Cookie", cookie.String())
	return w
}

func SetAccessTokenFastHttpCookie(requestCtx *fasthttp.RequestCtx, tokenString string, opts ...bool) {
	cookie := &fasthttp.Cookie{}
	cookie.SetKey("access_token")
	cookie.SetValue(tokenString)
	cookie.SetHTTPOnly(true)
	cookie.SetPath("/")
	if domainName != "" {
		cookie.SetDomain(domainName)
	}
	if secureToken {
		cookie.SetSecure(true)
	}
	if opts != nil && len(opts) > 0 {
		if opts[0] {
			cookie.SetExpire(time.Now().Add(time.Hour * tokenDuration))
			rememberMeCookie := &fasthttp.Cookie{}
			rememberMeCookie.SetValue("true")
			rememberMeCookie.SetKey("remember_me")
			rememberMeCookie.SetHTTPOnly(true)
			rememberMeCookie.SetPath("/")
			rememberMeCookie.SetExpire(time.Now().Add(time.Hour * tokenDuration))
			if domainName != "" {
				rememberMeCookie.SetDomain(domainName)
			}
			if secureToken {
				rememberMeCookie.SetSecure(true)
			}
			requestCtx.Response.Header.SetCookie(rememberMeCookie)
		}
	}
	requestCtx.Response.Header.SetCookie(cookie)
}

func RefreshAccessTokenCookie(w http.ResponseWriter, req *http.Request, userID int, role string) http.ResponseWriter {
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
			tokenString, tErr := jwtAuth.GenerateToken(userID, role)
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
				rememberMe := false
				rememberMeCookie, _ := req.Cookie("remember_me")
				if rememberMeCookie != nil {
					rememberMe, _ = strconv.ParseBool(rememberMeCookie.Value)
				}
				if rememberMe {
					cookie.Expires = time.Now().Add(time.Hour * tokenDuration)
					rememberMeCookie.Value = "true"
					rememberMeCookie.HttpOnly = true
					rememberMeCookie.Path = "/"
					rememberMeCookie.Expires = time.Now().Add(time.Hour * tokenDuration)
					if domainName != "" {
						rememberMeCookie.Domain = domainName
					}
					if secureToken {
						rememberMeCookie.Secure = true
					}
					w.Header().Add("Set-Cookie", rememberMeCookie.String())
				}
				w.Header().Add("Set-Cookie", cookie.String())
				w.Header().Add("Set-Cookie", tokenWasRefurbishedCookie.String())
			}
		}
	}
	return w
}

func RefreshAccessTokenFastHttpCookie(requestCtx *fasthttp.RequestCtx, userID int, role string) {
	tokenWasRefurbished := requestCtx.Request.Header.Cookie("token_was_refurbished")
	refresh := true
	if tokenWasRefurbished != nil {
		if r, parseErr := strconv.ParseBool(string(tokenWasRefurbished)); parseErr == nil {
			refresh = !r
		}
	}
	if refresh {
		cookieValue := requestCtx.Request.Header.Cookie("access_token")
		if cookieValue != nil {
			cookie := &fasthttp.Cookie{}
			cookie.SetKey("access_token")
			jwtAuth := GetJWTAuth()
			tokenString, tErr := jwtAuth.GenerateToken(userID, role)
			if tErr == nil {
				tokenWasRefurbishedCookie := &fasthttp.Cookie{}
				tokenWasRefurbishedCookie.SetKey("token_was_refurbished")
				tokenWasRefurbishedCookie.SetValue("true")
				tokenWasRefurbishedCookie.SetHTTPOnly(true)
				tokenWasRefurbishedCookie.SetPath("/")
				cookie.SetValue(tokenString)
				cookie.SetHTTPOnly(true)
				cookie.SetPath("/")
				if domainName != "" {
					cookie.SetDomain(domainName)
					tokenWasRefurbishedCookie.SetDomain(domainName)
				}
				if secureToken {
					tokenWasRefurbishedCookie.SetSecure(true)
					cookie.SetSecure(true)
				}
				rememberMe := false
				rememberMeCookieValue := requestCtx.Request.Header.Cookie("remember_me")
				if rememberMeCookieValue != nil {
					rememberMe, _ = strconv.ParseBool(string(rememberMeCookieValue))
				}
				if rememberMe {
					rememberMeCookie := &fasthttp.Cookie{}
					rememberMeCookie.SetKey("remember_me")
					cookie.SetExpire(time.Now().Add(time.Hour * tokenDuration))
					rememberMeCookie.SetValue("true")
					rememberMeCookie.SetHTTPOnly(true)
					rememberMeCookie.SetPath("/")
					rememberMeCookie.SetExpire(time.Now().Add(time.Hour * tokenDuration))
					if domainName != "" {
						rememberMeCookie.SetDomain(domainName)
					}
					if secureToken {
						rememberMeCookie.SetSecure(true)
					}
					requestCtx.Response.Header.SetCookie(rememberMeCookie)
				}
				requestCtx.Response.Header.SetCookie(cookie)
				requestCtx.Response.Header.SetCookie(tokenWasRefurbishedCookie)
			}
		}
	}
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
	cookie.Expires, _ = time.Parse(http.TimeFormat, http.TimeFormat)
	w.Header().Add("Set-Cookie", cookie.String())
	cookie.Name = "remember_me"
	cookie.Value = "false"
	w.Header().Add("Set-Cookie", cookie.String())
	return w
}
func RemoveAccessTokenFastHttpCookie(requestCtx *fasthttp.RequestCtx) {
	tokenCookie := &fasthttp.Cookie{}
	rememberMeCookie := &fasthttp.Cookie{}
	tokenCookie.SetKey("access_token")
	tokenCookie.SetValue("deleted")
	tokenCookie.SetHTTPOnly(true)
	tokenCookie.SetPath("/")
	rememberMeCookie.SetKey("remember_me")
	rememberMeCookie.SetValue("false")
	rememberMeCookie.SetHTTPOnly(true)
	rememberMeCookie.SetPath("/")
	if domainName != "" {
		tokenCookie.SetDomain(domainName)
		rememberMeCookie.SetDomain(domainName)
	}
	if secureToken {
		tokenCookie.SetSecure(true)
		rememberMeCookie.SetSecure(true)
	}
	expirationDate, _ := time.Parse(http.TimeFormat, http.TimeFormat)
	tokenCookie.SetExpire(expirationDate)
	rememberMeCookie.SetExpire(expirationDate)
	requestCtx.Response.Header.SetCookie(tokenCookie)
	requestCtx.Response.Header.SetCookie(rememberMeCookie)
}

func getTokenRemainingValidity(timestamp int64) int {
	tm := time.Unix(timestamp, 0)
	remainder := tm.Sub(time.Now())
	if remainder > 0 {
		return int(remainder.Seconds())
	}
	return -1
}
