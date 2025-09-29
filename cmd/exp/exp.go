package main

import (
	"log"
	"os"

	"github.com/wneessen/go-mail"
)

func main() {

	from := "test@test.com"
	to := "rahul@test.com"
	subject := "this is test mail"
	plaintext := "This is the body of the mail"
	html := "<h1>Hello there!</h1>"

	msg := mail.NewMsg()
	msg.From(from)
	msg.To(to)
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, plaintext)
	msg.AddAlternativeString(mail.TypeTextHTML, html)
	msg.WriteTo(os.Stdout)

	dialer, err := mail.NewClient(Host,
		mail.WithPort(Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(Username),
		mail.WithPassword(Password),
	)
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	err = dialer.DialAndSend(msg)
	if err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}

}
