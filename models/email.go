package models

import (
	"fmt"

	"github.com/wneessen/go-mail"
)

const (
	DefaultSender = "support@lenslocked.com"
)

type Email struct {
	From      string
	To        string
	Subject   string
	Plaintext string
	HTML      string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type EmailService struct {
	//DefaultSender is used as the default sender when one isn't provided for an
	//email. This is also used in functions where the email is a predetermined,
	//like the forgotten password email.
	DefaultSender string

	//unexported fields
	dialer *mail.Client
}

func NewEmailService(config SMTPConfig) (*EmailService, error) {
	client, err := mail.NewClient(config.Host,
		mail.WithPort(config.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.Username),
		mail.WithPassword(config.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	es := EmailService{
		DefaultSender: DefaultSender,
		dialer:        client,
	}
	return &es, nil
}

// Set the Client into dialer in NewEmailService, then
// here set new mail msg object over that dialer(EmailService.dialer)
// then DialAndSend over the Dialer type
func (es *EmailService) Send(email Email) error {
	msg := mail.NewMsg()
	// msg.To(email.To)
	// es.setFrom(msg, email)
	// Handle errors from From and To methods
	if err := msg.To(email.To); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}

	// Set the From field to a default value if it's not set in the Email
	if err := es.setFrom(msg, email); err != nil {
		return fmt.Errorf("failed to set From address: %w", err)
	}

	msg.Subject(email.Subject)
	//This switch is created for:
	//if only one of plainText or Html is set then only show one thing in the body
	switch {
	case email.Plaintext != "" && email.HTML != "":
		msg.SetBodyString(mail.TypeTextPlain, email.Plaintext)
		msg.AddAlternativeString(mail.TypeTextHTML, email.HTML)
	case email.Plaintext != "":
		msg.SetBodyString(mail.TypeTextPlain, email.Plaintext)
	case email.HTML != "":
		msg.SetBodyString(mail.TypeTextHTML, email.HTML)
	}

	err := es.dialer.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}

func (es *EmailService) ForgotPassword(to, resetURL string) error {
	email := Email{
		Subject:   "Reset your password",
		To:        to,
		Plaintext: "To reset your passoword, please visit the following link: " + resetURL,
		HTML: `<p>To reset your password, please visit the following link: <a
		href="` + resetURL + `">` + resetURL + `</a></p>`,
	}
	err := es.Send(email)
	if err != nil {
		return fmt.Errorf("forgot password email: %w", err)
	}
	return nil
}

func (es *EmailService) setFrom(msg *mail.Msg, email Email) error {
	var from string
	switch {
	case email.From != "":
		from = email.From
	case es.DefaultSender != "":
		from = es.DefaultSender
	default:
		from = DefaultSender
	}
	return msg.From(from)
}

// Close closes the client connection
func (es *EmailService) Close() error {
	if es.dialer != nil {
		return es.dialer.Close()
	}
	return nil
}
