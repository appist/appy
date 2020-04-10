package mailer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/CloudyKit/jet"
	"github.com/appist/appy/support"
	"github.com/appist/appy/view"
	"github.com/jordan-wright/email"
)

// Engine provides the capability to parse/render email template and send it
// out via SMTP protocol.
type Engine struct {
	config     *support.Config
	deliveries []*Mail
	i18n       *support.I18n
	previews   map[string]*Mail
	smtpAddr   string
	smtpAuth   smtp.Auth
	viewEngine *view.Engine
}

// NewEngine initializes Mailer instance.
func NewEngine(asset *support.Asset, config *support.Config, i18n *support.I18n, logger *support.Logger, viewFuncs map[string]interface{}) *Engine {
	ve := view.NewEngine(asset, config, logger)
	ve.SetGlobalFuncs(viewFuncs)

	mailer := &Engine{
		config:   config,
		i18n:     i18n,
		previews: map[string]*Mail{},
		smtpAddr: config.MailerSMTPAddr,
		smtpAuth: smtp.PlainAuth(
			config.MailerSMTPPlainAuthIdentity,
			config.MailerSMTPPlainAuthUsername,
			config.MailerSMTPPlainAuthPassword,
			config.MailerSMTPPlainAuthHost,
		),
		viewEngine: ve,
	}

	return mailer
}

// AddPreview add the mail HTML/text template preview.
func (e *Engine) AddPreview(mail *Mail) {
	e.previews[mail.Template] = mail
}

// ComposeEmail constructs the HTML/text content and transforms mailer.Mail
// into email.Email.
func (e *Engine) ComposeEmail(mail *Mail) (*email.Email, error) {
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
		mail.Locale = e.config.I18nDefaultLocale
	}

	subject := e.i18n.T(mail.Subject, mail.Locale)
	if subject != "" {
		email.Subject = subject
	}

	if mail.Headers == nil {
		email.Headers = textproto.MIMEHeader{}
	}

	html, err := e.content(mail.Locale, mail.Template+".html", "html", mail.TemplateData)
	if err != nil {
		return nil, err
	}
	mail.HTML = string(html)
	email.HTML = html

	text, err := e.content(mail.Locale, mail.Template+".txt", "txt", mail.TemplateData)
	if err != nil {
		return nil, err
	}
	mail.Text = string(text)
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

// Deliver sends the email via SMTP protocol without TLS.
func (e *Engine) Deliver(mail *Mail) error {
	email, err := e.ComposeEmail(mail)
	if err != nil {
		return err
	}

	if e.config.AppyEnv == "test" {
		e.deliveries = append(e.deliveries, mail)
		return nil
	}

	return email.SendWithTLS(e.smtpAddr, e.smtpAuth, &tls.Config{})
}

// Deliveries returns the Mail array which is used for unit test with
// APPY_ENV=test.
func (e *Engine) Deliveries() []*Mail {
	return e.deliveries
}

// Previews returns all the templates preview.
func (e *Engine) Previews() map[string]*Mail {
	return e.previews
}

func (e *Engine) content(locale, name, ext string, obj interface{}) ([]byte, error) {
	set := e.viewEngine.HTMLSet()

	if ext == "txt" {
		set = e.viewEngine.TxtSet()
	}

	set.AddGlobal("t", func(key string, args ...interface{}) string {
		var (
			tplCount  int
			tplData   support.H
			tplLocale string
		)

		for _, arg := range args {
			switch v := arg.(type) {
			case float64:
				tplCount = int(v)
			case string:
				if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
					_ = json.Unmarshal([]byte(v), &tplData)
					continue
				}

				tplLocale = v
			}
		}

		if tplCount != 0 {
			args = append(args, tplCount)
		}

		if tplData != nil {
			args = append(args, tplData)
		}

		if tplLocale == "" {
			args = append(args, locale)
		}

		return e.i18n.T(key, args...)
	})

	t, err := set.GetTemplate(name)
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
