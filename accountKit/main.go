package accountKit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/parnurzeal/gorequest"
)

var (
	apiVersion           string = "v1.0"
	appID                string
	appSecret            string
	meEndpointBaseURL    string
	tokenExchangeBaseURL string
)

func Init(AppId string, AppSecret string, opts ...string) {
	if opts != nil && len(opts) > 0 {
		apiVersion = opts[0]
	}
	appID = AppId
	appSecret = AppSecret
	meEndpointBaseURL = "https://graph.accountkit.com/" + apiVersion + "/me"
	tokenExchangeBaseURL = "https://graph.accountkit.com/" + apiVersion + "/access_token"
}

func ExchangeToken(authCode string) (string, error) {
	if authCode != "" {
		tokenURL, _ := url.Parse(tokenExchangeBaseURL)
		query := tokenURL.Query()
		query.Set("grant_type", "authorization_code")
		query.Set("code", authCode)
		query.Set("access_token", "AA|"+appID+"|"+appSecret)
		tokenURL.RawQuery = query.Encode()
		request := gorequest.New()
		_, body, errs := request.Get(tokenURL.String()).End()
		if errs != nil && len(errs) > 0 {
			return "", errs[0]
		}
		resp := map[string]string{}
		if err := json.Unmarshal([]byte(body), &resp); err == nil {
			if accessToken, ok := resp["access_token"]; ok {
				return accessToken, nil
			}
		}
		return "", errors.New("invalid_authorization_code")
	}
	return "", errors.New("authorization_code_was_missed")
}

func AccountDetails(accessToken string) (map[string]string, error) {
	if accessToken != "" {
		URL, _ := url.Parse(meEndpointBaseURL)
		query := URL.Query()
		query.Set("access_token", accessToken)
		query.Set("appsecret_proof", generateAppSecretProof(accessToken))
		URL.RawQuery = query.Encode()
		request := gorequest.New()
		_, body, errs := request.Get(URL.String()).End()
		if errs != nil && len(errs) > 0 {
			return nil, errs[0]
		}
		resp := map[string]string{}
		if err := json.Unmarshal([]byte(body), &resp); err == nil {
			return resp, nil
		}
		return nil, errors.New("invalid_access_token")
	}
	return nil, errors.New("access_token_was_missed")
}

func IsPhoneNumberOwner(phoneNumber string, authCode string, opts ...bool) bool {
	national := false
	if opts != nil && len(opts) > 0 {
		national = opts[0]
	}
	if accessToken, err1 := ExchangeToken(authCode); err1 == nil {
		if accountDetails, err2 := AccountDetails(accessToken); err2 == nil {
			if accountDetails != nil {
				if national {
					if nationalNumber, ok := accountDetails["national_number"]; ok {
						if nationalNumber == phoneNumber {
							return true
						}
					}
				} else {
					if number, ok := accountDetails["number"]; ok {
						if number == phoneNumber {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func generateAppSecretProof(accessToken string) string {
	asp := hmac.New(sha256.New, []byte(appSecret))
	asp.Write([]byte(accessToken))
	return hex.EncodeToString(asp.Sum(nil))
}
