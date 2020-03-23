package mailer

import "github.com/appist/appy"

func newMailPreview() appy.Mail {
	return appy.Mail{
		From:    "support@appy.org",
		To:      []string{"a@appy.org"},
		ReplyTo: []string{"b@appy.org"},
		Cc:      []string{"c@appy.org"},
		Bcc:     []string{"d@appy.org"},
	}
}

func newMail() appy.Mail {
	return appy.Mail{
		From: "support@appy.org",
	}
}
