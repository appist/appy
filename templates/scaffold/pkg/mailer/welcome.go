package mailer

import (
	"{{.Project.Name}}/pkg/app"

	"github.com/appist/appy"
)

func init() {
	app.Mailer.AddPreview(welcomePreview())
}

func welcome() appy.Mail {
	mail := newMail()
	mail.Subject = "mailers.welcome.subject"
	mail.Template = "mailers/welcome"
	mail.TemplateData = appy.H{}

	return mail
}

func welcomePreview() appy.Mail {
	mail := newMailPreview()
	mail.Subject = "mailers.welcome.subject"
	mail.Template = "mailers/welcome"
	mail.TemplateData = appy.H{}

	return mail
}
