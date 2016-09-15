package antiCSRF

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/nehmeroumani/nrgo/clean"
)

const (
	tokenRequestHeader string = "X-CSRF-TOKEN"
	tokenFieldName     string = "csrf_token"
	tokenCookieName    string = "csrf_base_token"
)

var (
	tokenSalt   string
	tokenLength int
	domainName  string
	secureToken bool
)

func Init(TokenLength int, DomainName string, opts ...interface{}) {
	tokenLength = TokenLength
	domainName = DomainName
	if opts != nil {
		if len(opts) > 0 {
			secureToken = opts[0].(bool)
			if len(opts) > 1 {
				tokenSalt = opts[1].(string)
			}
		}
	}
}

func NewCSRFToken(opts ...interface{}) *CSRFToken {
	token := &CSRFToken{}
	if opts != nil {
		if len(opts) > 0 {
			token.Encrypted = opts[0].(bool)
			if token.Encrypted && tokenSalt == "" {
				token.Encrypted = false
			}
		}
	}
	var err error
	randBytes, _ := generateRandomBytes(tokenLength)
	token.RealToken = string(randBytes)
	if token.Encrypted {
		token.RealToken, err = encrypt([]byte(tokenSalt), token.RealToken)
		if err != nil {
			clean.Error(err)
		}
	}
	return token
}

type CSRFToken struct {
	RealToken     string
	Encrypted     bool
	MaskedToken   string
	UnmaskedToken string
}

func (this *CSRFToken) WithMask() string {
	otp, err := generateRandomBytes(tokenLength)
	if err != nil {
		return ""
	}
	this.MaskedToken = base64.StdEncoding.EncodeToString(append(otp, xorToken(otp, []byte(this.RealToken))...))
	return this.MaskedToken
}

func (this *CSRFToken) Unmask(issued string) string {
	issuedBytes := []byte(issued)
	if len(issuedBytes) != tokenLength*2 {
		return ""
	}

	otp := issuedBytes[tokenLength:]
	masked := issuedBytes[:tokenLength]

	if token := xorToken(otp, masked); token != nil {
		this.UnmaskedToken = string(token)
		return this.UnmaskedToken
	}
	return ""
}

func (this *CSRFToken) IsValidRequestToken() bool {
	var err error
	realToken := this.RealToken
	if this.Encrypted {
		if _, err = decrypt([]byte(tokenSalt), realToken); err != nil {
			return false
		}
	}
	a := []byte(realToken)
    if this.UnmaskedToken == ""{
        this.Unmask(this.MaskedToken)
    }
	b := []byte(this.UnmaskedToken)
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(a, b) == 1
}

func (this *CSRFToken) SetCookie(w http.ResponseWriter) http.ResponseWriter {
	cookie := http.Cookie{}
	cookie.Name = tokenCookieName
	cookie.Value = this.RealToken
	cookie.HttpOnly = true
	cookie.Path = "/"
	if domainName != "" {
		cookie.Domain = domainName
	}
	if secureToken {
		cookie.Secure = true
	}
	w.Header().Add("Set-Cookie", cookie.String())
	return w
}

func (this *CSRFToken) HTMLInput() string {
	if this.MaskedToken == "" {
		this.WithMask()
	}
	input := fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
		tokenFieldName, this.MaskedToken)
	return input
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// err == nil only if len(b) == n
	if err != nil {
		return nil, err
	}

	return b, nil
}

func xorToken(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	res := make([]byte, n)

	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}

	return res
}

func GetRequestCSRFToken(r *http.Request) *CSRFToken {
	// 1. Check the HTTP header first.
	requestToken := r.Header.Get(tokenRequestHeader)

	// 2. Fall back to the POST (form) value.
	if requestToken == "" {
		requestToken = r.PostFormValue(tokenFieldName)
	}

	// 3. Finally, fall back to the multipart form (if set).
	if requestToken == "" && r.MultipartForm != nil {
		vals := r.MultipartForm.Value[tokenFieldName]

		if len(vals) > 0 {
			requestToken = vals[0]
		}
	}

	// Decode the "issued" (pad + masked) token sent in the request. Return a
	// nil byte slice on a decoding error (this will fail upstream).
	decodedRequestToken, _ := base64.StdEncoding.DecodeString(requestToken)

	realTokenCookie, err := r.Cookie(tokenCookieName)
	if err != nil {
		return nil
	}
	csrfToken := &CSRFToken{}
	csrfToken.RealToken = realTokenCookie.Value
	csrfToken.MaskedToken = string(decodedRequestToken)
	return csrfToken
}