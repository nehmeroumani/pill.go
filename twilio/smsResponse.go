package twilio

import "time"

// SmsResponse is returned after a text/sms message is posted to Twilio
type SmsResponse struct {
	Sid         string   `json:"sid"`
	DateCreated string   `json:"date_created"`
	DateUpdate  string   `json:"date_updated"`
	DateSent    string   `json:"date_sent"`
	AccountSid  string   `json:"account_sid"`
	To          string   `json:"to"`
	From        string   `json:"from"`
	MediaUrl    string   `json:"media_url"`
	Body        string   `json:"body"`
	Status      string   `json:"status"`
	Direction   string   `json:"direction"`
	ApiVersion  string   `json:"api_version"`
	Price       *float32 `json:"price,omitempty"`
	Url         string   `json:"uri"`
}

// Returns SmsResponse.DateCreated as a time.Time object
// instead of a string.
func (sms *SmsResponse) DateCreatedAsTime() (time.Time, error) {
	return time.Parse(time.RFC1123Z, sms.DateCreated)
}

// Returns SmsResponse.DateUpdate as a time.Time object
// instead of a string.
func (sms *SmsResponse) DateUpdateAsTime() (time.Time, error) {
	return time.Parse(time.RFC1123Z, sms.DateUpdate)
}

// Returns SmsResponse.DateSent as a time.Time object
// instead of a string.
func (sms *SmsResponse) DateSentAsTime() (time.Time, error) {
	return time.Parse(time.RFC1123Z, sms.DateSent)
}
