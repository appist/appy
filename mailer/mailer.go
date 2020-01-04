package mailer

import (
	"bytes"
	"crypto/tls"
	"net/smtp"
	"net/textproto"

	"github.com/CloudyKit/jet"
	"github.com/appist/appy/support"
	"github.com/jordan-wright/email"
)

type (
	// Mailer provides the capability to parse/render email template and send it out via SMTP protocol.
	Mailer struct {
		addr       string
		config     *support.Config
		i18n       *support.I18n
		plainAuth  smtp.Auth
		previews   map[string]Email
		viewEngine *support.ViewEngine
	}

	// Email defines the email headers/body/attachments.
	Email struct {
		From         string
		To           []string
		ReplyTo      []string
		Bcc          []string
		Cc           []string
		Sender       string
		Subject      string
		Headers      textproto.MIMEHeader
		Template     string
		TemplateData interface{}
		Attachments  []string
		ReadReceipt  []string
		Locale       string
	}
)

// NewMailer initializes Mailer instance.
func NewMailer(assets *support.Assets, config *support.Config, i18n *support.I18n) *Mailer {
	mailer := &Mailer{
		addr:   config.MailerSMTPAddr,
		config: config,
		i18n:   i18n,
		plainAuth: smtp.PlainAuth(
			config.MailerSMTPPlainAuthIdentity,
			config.MailerSMTPPlainAuthUsername,
			config.MailerSMTPPlainAuthPassword,
			config.MailerSMTPPlainAuthHost,
		),
		previews:   map[string]Email{},
		viewEngine: support.NewViewEngine(assets),
	}

	return mailer
}

// AddPreview add the mail HTML/text template preview.
func (m *Mailer) AddPreview(mail Email) {
	m.previews[mail.Template] = mail
}

// Previews returns all the templates preview.
func (m *Mailer) Previews() map[string]Email {
	return m.previews
}

// Send sends the email via SMTP protocol without TLS.
func (m *Mailer) Send(mail Email) error {
	email, err := m.ComposeEmail(mail)
	if err != nil {
		return err
	}

	return email.Send(m.addr, m.plainAuth)
}

// SendWithTLS sends the email via SMTP protocol with TLS.
func (m *Mailer) SendWithTLS(mail Email) error {
	email, err := m.ComposeEmail(mail)
	if err != nil {
		return err
	}

	// TODO: Figure out how to easily configure TLS with appy's configuration.
	tls := &tls.Config{}

	return email.SendWithTLS(m.addr, m.plainAuth, tls)
}

func (m *Mailer) content(locale, name string, obj interface{}) ([]byte, error) {
	m.viewEngine.AddGlobal("t", func(key string, args ...interface{}) string {
		var tplLocale string
		for _, arg := range args {
			switch v := arg.(type) {
			case string:
				tplLocale = v
			}
		}

		if tplLocale == "" {
			args = append(args, locale)
		}

		return m.i18n.T(key, args...)
	})

	t, err := m.viewEngine.GetTemplate(name)
	if err != nil {
		return nil, err
	}

	var w bytes.Buffer
	vars := make(jet.VarMap)
	if err = t.Execute(&w, vars, obj); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// ComposeEmail returns the email instance with html/txt templates.
func (m *Mailer) ComposeEmail(mail Email) (*email.Email, error) {
	email := &email.Email{
		From:        mail.From,
		To:          mail.To,
		ReplyTo:     mail.ReplyTo,
		Bcc:         mail.Bcc,
		Cc:          mail.Cc,
		Sender:      mail.Sender,
		Subject:     mail.Subject,
		ReadReceipt: mail.ReadReceipt,
	}

	if mail.Locale == "" {
		mail.Locale = m.config.I18nDefaultLocale
	}

	subject := m.i18n.T(mail.Subject, mail.Locale)
	if subject != "" {
		email.Subject = subject
	}

	if mail.Headers == nil {
		email.Headers = textproto.MIMEHeader{}
	}

	html, err := m.content(mail.Locale, mail.Template+".html", mail.TemplateData)
	if err != nil {
		return nil, err
	}
	email.HTML = html

	text, err := m.content(mail.Locale, mail.Template+".txt", mail.TemplateData)
	if err != nil {
		return nil, err
	}
	email.Text = text

	if mail.Attachments != nil {
		for _, attachment := range mail.Attachments {
			_, err := email.AttachFile(attachment)

			if err != nil {
				return nil, err
			}
		}
	}

	return email, nil
}
