package accountKit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/nehmeroumani/pill.go/clean"
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

type AccountDetails struct {
	ID    string `json:"id"`
	Phone *Phone `json:"phone"`
}

type Phone struct {
	Number         string `json:"number"`
	NationalNumber string `json:"national_number"`
	CountryPrefix  string `json:"country_prefix"`
}

func NewAccountDetails() *AccountDetails {
	accountDetails := &AccountDetails{}
	accountDetails.Phone = &Phone{}
	return accountDetails
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
		resp := map[string]interface{}{}
		if err := json.Unmarshal([]byte(body), &resp); err == nil {
			if accessToken, ok := resp["access_token"]; ok {
				return accessToken.(string), nil
			}
		} else {
			clean.Error(err)
		}
		return "", errors.New("invalid_authorization_code")
	}
	return "", errors.New("authorization_code_was_missed")
}

func GetAccountDetails(accessToken string) (*AccountDetails, error) {
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
		resp := NewAccountDetails()
		if err := json.Unmarshal([]byte(body), &resp); err == nil {
			return resp, nil
		} else {
			clean.Error(err)
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
		if accountDetails, err2 := GetAccountDetails(accessToken); err2 == nil {
			if accountDetails != nil {
				if national {
					if accountDetails.Phone.NationalNumber == phoneNumber {
						return true
					} else if strings.Replace(accountDetails.Phone.NationalNumber, "+", "", -1) == phoneNumber {
						return true
					}
				} else {
					if accountDetails.Phone.Number == phoneNumber {
						return true
					} else if strings.Replace(accountDetails.Phone.Number, "+", "", -1) == phoneNumber {
						return true
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
