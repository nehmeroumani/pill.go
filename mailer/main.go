package mailer

import (
	"bytes"
	"errors"
	"time"

	"github.com/nehmeroumani/pill.go/clean"
	"github.com/nehmeroumani/pill.go/templates"

	"gopkg.in/gomail.v2"
)

//MAILER CONFIGURATIONS
var host string
var port int
var email string
var password string
var senderName string

var ch = make(chan *gomail.Message)

func Init(Host string, Port int, SenderName string, Email string, Password string) {
	host = Host
	port = Port
	senderName = SenderName
	email = Email
	password = Password
	go run()
}

func run() {
	d := gomail.NewDialer(host, port, email, password)

	var s gomail.SendCloser
	var err error
	open := false
	for {
		select {
		case m, ok := <-ch:
			if !ok {
				return
			}
			if !open {
				if s, err = d.Dial(); err != nil {
					clean.Error(err)
					return
				}
				open = true
			}
			if err := gomail.Send(s, m); err != nil {
				clean.Error(err)
			}
		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				if err := s.Close(); err != nil {
					clean.Error(err)
					return
				}
				open = false
			}
		}
	}
}

func Send(to []string, subject string, templateName string, data interface{}, attachments ...string) {
	m := gomail.NewMessage()
	m.SetHeader("From", senderName+" <"+email+">")
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	if attachments != nil && len(attachments) > 0 {
		for _, attachment := range attachments {
			m.Attach(attachment)
		}
	}
	var body bytes.Buffer
	tmpl := templates.GetTemplate(templateName)
	if tmpl == nil {
		clean.Error(errors.New("Template '" + templateName + "' not exist"))
		return
	}
	if err := tmpl.Execute(&body, data); err != nil {
		clean.Error(err)
		return
	}
	content := body.String()
	m.SetBody("text/html", content)
	ch <- m
}
