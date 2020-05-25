package mailer

import (
	"{{.projectName}}/pkg/app"

	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
)

func init() {
	app.Mailer.AddPreview(welcomePreview())
}

func welcome() *mailer.Mail {
	mail := newMail()
	mail.Subject = "mailers.welcome.subject"
	mail.Template = "mailers/welcome"
	mail.TemplateData = support.H{}

	return mail
}

func welcomePreview() *mailer.Mail {
	mail := newMailPreview()
	mail.Subject = "mailers.welcome.subject"
	mail.Template = "mailers/welcome"
	mail.TemplateData = support.H{}

	return mail
}
