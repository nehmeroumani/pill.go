package antiCSRF

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/nehmeroumani/pill.go/clean"
)

const (
	tokenRequestHeader string = "X-CSRF-TOKEN"
	tokenFieldName     string = "csrf_token"
	tokenCookieName    string = "csrf_base_token"
)

var (
	encryptionKey      string
	tokenLength    int
	domainName     string
	encryptedToken bool
	secureToken    bool
	safeMethods    = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
)

func Init(TokenLength int, DomainName string, opts ...interface{}) {
	tokenLength = TokenLength
	domainName = DomainName
	if opts != nil {
		if len(opts) > 0 {
			secureToken = opts[0].(bool)
			if len(opts) > 1 {
				encryptionKey = opts[1].(string)
			}
		}
	}
	if encryptionKey != "" {
		encryptedToken = true
	}
}

func NewCSRFToken() *CSRFToken {
	token := &CSRFToken{}
	randBytes, _ := generateRandomBytes(tokenLength)
	token.RealToken = string(randBytes)
	return token
}

type CSRFToken struct {
	RealToken     string
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
	a := []byte(this.RealToken)
	if this.UnmaskedToken == "" {
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
	if encryptedToken {
		var err error
		cookie.Value, err = encrypt([]byte(encryptionKey), this.RealToken)
		if err != nil {
			clean.Error(err)
			return w
		}
	} else {
		cookie.Value = base64.StdEncoding.EncodeToString([]byte(this.RealToken))
	}
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

	realTokenCookie, err1 := r.Cookie(tokenCookieName)
	if err1 != nil {
		clean.Error(err1)
		return nil
	}
	realToken := realTokenCookie.Value
	if !encryptedToken {
		decodedRealToken, err2 := base64.StdEncoding.DecodeString(realTokenCookie.Value)
		if err2 != nil {
			clean.Error(err2)
			return nil
		}
		realToken = string(decodedRealToken)
	}
	if encryptedToken {
		var err3 error
		if realToken, err3 = decrypt([]byte(encryptionKey), realToken); err3 != nil {
			clean.Error(err3)
			return nil
		}
	}
	csrfToken := &CSRFToken{}
	csrfToken.RealToken = realToken
	csrfToken.MaskedToken = string(decodedRequestToken)
	return csrfToken
}

func contains(vals []string, s string) bool {
	for _, v := range vals {
		if v == s {
			return true
		}
	}

	return false
}

func IsSafeMethod(method string) bool {
	if contains(safeMethods, method) {
		return true
	}
	return false
}
