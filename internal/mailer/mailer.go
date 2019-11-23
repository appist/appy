package mailer

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strings"

	appyhttp "github.com/appist/appy/internal/http"
	appysupport "github.com/appist/appy/internal/support"
	"github.com/jordan-wright/email"
)

type (
	// Mailer provides the capability to parse/render email template and send it out via SMTP protocol.
	Mailer struct {
		addr      string
		plainAuth smtp.Auth
		previews  map[string]Mail
		server    *appyhttp.Server
	}

	// Mail defines the email headers/body/attachments.
	Mail struct {
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
	}
)

// NewMailer initializes Mailer instance.
func NewMailer(c *appysupport.Config, l *appysupport.Logger, s *appyhttp.Server) *Mailer {
	mailer := &Mailer{
		addr: c.MailerAddr,
		plainAuth: smtp.PlainAuth(
			c.MailerPlainAuthIdentity,
			c.MailerPlainAuthUsername,
			c.MailerPlainAuthPassword,
			c.MailerPlainAuthHost,
		),
		previews: map[string]Mail{},
		server:   s,
	}

	if appysupport.IsDebugBuild() {
		s.HTMLRenderer().AddFromString("mailer/preview", previewTpl())

		s.Router().GET(s.Config().MailerPreviewPath, func(ctx *appyhttp.Context) {
			ctx.HTML(http.StatusOK, "mailer/preview", appyhttp.H{
				"previewPath": s.Config().MailerPreviewPath,
				"previews":    mailer.previews,
				"title":       "Mailer Preview",
			})
		})

		s.Router().GET(s.Config().MailerPreviewPath+"/preview", func(ctx *appyhttp.Context) {
			contentType := "text/html"
			name := ctx.Query("name")
			preview, exists := mailer.previews[name]
			if !exists {
				ctx.AbortWithStatus(http.StatusNotFound)
				return
			}

			data, err := mailer.Content("html", name, preview.TemplateData)
			if err != nil {
				ctx.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			if ctx.Query("ext") == "txt" {
				contentType = "text/plain"
				data, err = mailer.Content("txt", name, preview.TemplateData)
				if err != nil {
					ctx.AbortWithError(http.StatusInternalServerError, err)
					return
				}
			}

			ctx.Writer.Header().Del(http.CanonicalHeaderKey("x-frame-options"))
			ctx.Data(http.StatusOK, contentType, data)
		})
	}

	return mailer
}

// Preview sets up the preview for the mail HTML/text template.
func (m *Mailer) Preview(mail Mail) {
	m.previews[mail.Template] = mail
}

// Send delivers the email via SMTP protocol without TLS.
func (m Mailer) Send(mail Mail) error {
	email, err := m.composeEmail(mail)
	if err != nil {
		return err
	}

	return email.Send(m.addr, m.plainAuth)
}

// SendWithTLS delivers the email via SMTP protocol with TLS.
func (m Mailer) SendWithTLS(mail Mail, tls *tls.Config) error {
	email, err := m.composeEmail(mail)
	if err != nil {
		return err
	}

	return email.SendWithTLS(m.addr, m.plainAuth, tls)
}

// Content returns the content for the named email template.
func (m Mailer) Content(ext, name string, data interface{}) ([]byte, error) {
	var dataCopy interface{}
	appysupport.DeepClone(dataCopy, data)

	if dataCopy == nil {
		dataCopy = appyhttp.H{}
	}

	if _, ok := dataCopy.(appyhttp.H)["layout"]; !ok {
		dataCopy.(appyhttp.H)["layout"] = "mailer_default." + ext
	}

	recorder := httptest.NewRecorder()
	renderer := m.server.HTMLRenderer().Instance(name+"."+ext, dataCopy)

	if err := renderer.Render(recorder); err != nil {
		return nil, err
	}

	return recorder.Body.Bytes(), nil
}

func (m Mailer) composeEmail(mail Mail) (*email.Email, error) {
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

	if mail.Headers == nil {
		email.Headers = textproto.MIMEHeader{}
	}

	tpl := mail.Template
	if regexp.MustCompile(`\.html$`).Match([]byte(tpl)) {
		tpl = strings.TrimSuffix(tpl, ".html")
	} else if regexp.MustCompile(`\.txt$`).Match([]byte(tpl)) {
		tpl = strings.TrimSuffix(tpl, ".txt")
	}

	html, err := m.Content("html", mail.Template, mail.TemplateData)
	if err != nil {
		return nil, err
	}
	email.HTML = html

	text, err := m.Content("txt", mail.Template, mail.TemplateData)
	if err != nil {
		return nil, err
	}
	email.Text = text

	if mail.Attachments != nil {
		for _, attachment := range mail.Attachments {
			// TODO: Add external/internal file handling.
			email.AttachFile(attachment)
		}
	}

	return email, nil
}
