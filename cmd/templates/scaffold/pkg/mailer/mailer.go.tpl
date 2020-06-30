package mailer

import "github.com/appist/appy/mailer"

func newMailPreview() *mailer.Mail {
	return &mailer.Mail{
		From:    "support@appy.org",
		To:      []string{"a@appy.org"},
		ReplyTo: []string{"b@appy.org"},
		Cc:      []string{"c@appy.org"},
		Bcc:     []string{"d@appy.org"},
	}
}

func newMail() *mailer.Mail {
	return &mailer.Mail{
		From: "support@appy.org",
	}
}
