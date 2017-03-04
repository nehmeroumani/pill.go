package twilio

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nehmeroumani/pill.go/clean"
)

var accountSid, authToken, messagingServiceSid, from string

func Init(AccountSid, AuthToken, MessagingServiceSid, From string) {
	accountSid = AccountSid
	authToken = AuthToken
	messagingServiceSid = MessagingServiceSid
	from = From
}

func SendSms(to, body string) *SmsResponse {
	// Set initial variables
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"

	// Build out the data for our message
	v := url.Values{}
	v.Set("To", to)
	v.Set("From", from)
	v.Set("Body", body)
	v.Set("MessagingServiceSid", messagingServiceSid)
	rb := *strings.NewReader(v.Encode())

	// Create client
	client := newClient()

	req, _ := http.NewRequest("POST", urlStr, &rb)

	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	resp, _ := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			smsResponse := &SmsResponse{}
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			err := json.Unmarshal(bodyBytes, smsResponse)
			if err == nil {
				return smsResponse
			} else {
				clean.Error(err)
			}
		}
	}
	return nil

}

func newClient() *http.Client {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 30 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 60,
		Transport: netTransport,
	}
	return netClient
}
