// Package clients provides a client for the email service.
package clients

import (
	"context"
	"os"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/email_sending"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/shared"
)

func SendEmail(title string, body string) error {
	client := cloudflare.NewClient(option.WithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN")))

	_, err := client.EmailSending.Send(context.TODO(), email_sending.EmailSendingSendParams{
		AccountID: cloudflare.F(os.Getenv("CLOUDFLARE_ACCOUNT_ID")),
		From:      cloudflare.F[email_sending.EmailSendingSendParamsFromUnion](shared.UnionString(os.Getenv("FROM_EMAIL"))),
		To:        cloudflare.F[email_sending.EmailSendingSendParamsToUnion](shared.UnionString(os.Getenv("TO_EMAIL"))),
		Subject:   cloudflare.F(title),
		HTML:      cloudflare.F(body),
	})
	return err
}
