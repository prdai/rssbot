// Package clients provides a client for the email service.
package clients

import (
	"log/slog"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendEmail(title string, body string) {
	apiKey := os.Getenv("RESEND_API_KEY")

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{os.Getenv("TO_EMAIL")},
		Subject: title,
		Html:    body,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		slog.Error(err.Error())
	}
}
