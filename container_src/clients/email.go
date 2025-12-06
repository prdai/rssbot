// Package clients provides a client for the email service.
package clients

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

func SendEmail(title string, body string) {
	server := mail.NewSMTPClient()
	server.Host = "smtp.gmail.com"
	server.Port = 587
	server.Username = os.Getenv("EMAIL_ADDRESS")
	server.Password = os.Getenv("EMAIL_PASSWORD")
	server.Encryption = mail.Encryption(mail.AuthNone)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	smtpClient, err := server.Connect()
	if err != nil {
		slog.Error(err.Error())
	}
	email := mail.NewMSG()
	email.SetFrom(fmt.Sprintf("RSSbot <%s>", os.Getenv("EMAIL_ADDRESS")))
	email.AddTo(os.Getenv("TO_EMAIL"))
	email.SetSubject(title)
	email.SetBody(mail.TextHTML, body)
	err = email.Send(smtpClient)
	if err != nil {
		slog.Error(err.Error())
	}
}
