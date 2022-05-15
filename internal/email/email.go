package email

import (
	"errors"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"
)

type EmailSetting struct {
	ServiceProvider string
	Host            string
	Port            int
	User            string
	Password        string
	Recipients      []string
}

var (
	MissingSettings = errors.New("no email setting provided")
)

var e *EmailSetting

func ConfigEmailService(settings *EmailSetting) {
	e = settings
}

// TestEmailService send a test message to email address provided
func TestEmailService(email string) error {
	temp := e.Recipients
	e.Recipients = []string{email}
	err := SendEmail("Testing", "This is a testing message.")
	e.Recipients = temp
	return err
}

func SendEmail(subject string, body string, v ...interface{}) error {
	if e == nil {
		return MissingSettings
	}

	var auth smtp.Auth
	if e.ServiceProvider != "Outlook" && e.ServiceProvider != "outlook" {
		auth = smtp.PlainAuth("", e.User, e.Password, e.Host)
	} else {
		auth = OutlookLoginAuth(e.User, e.Password)
	}

	msg := buildMessage(e.User, &mailContent{
		To:      e.Recipients,
		Subject: subject,
		Body:    fmt.Sprintf(body, v...),
	})

	if err := send(auth, msg); err != nil {
		return fmt.Errorf("failed to send email: %v\nmassage to be sent:\n%v", err, msg)
	}

	return nil
}

func send(auth smtp.Auth, msg string) error {
	if e == nil {
		return MissingSettings
	}

	return smtp.SendMail(e.Host+":"+strconv.Itoa(e.Port), auth, e.User, e.Recipients, []byte(msg))
}

type mailContent struct {
	To      []string
	Cc      []string
	Subject string
	Body    string
}

func buildMessage(from string, content *mailContent) string {

	message := strings.Builder{}
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))

	if len(content.To) > 0 {
		message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(content.To, ";")))
	}

	if len(content.Cc) > 0 {
		message.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(content.Cc, ";")))
	}

	message.WriteString(fmt.Sprintf("Subject: %s\r\n", content.Subject))
	message.WriteString(fmt.Sprintf("\r\n%s\r\n", content.Body))

	return message.String()
}

type auth struct {
	username, password string
}

func OutlookLoginAuth(username, password string) smtp.Auth {
	return &auth{username, password}
}

func (a *auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown from Server")
		}
	}
	return nil, nil
}
