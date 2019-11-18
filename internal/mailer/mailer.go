package mailer

import (
	"net/http/httptest"

	appyhttp "github.com/appist/appy/internal/http"
	appysupport "github.com/appist/appy/internal/support"
	"github.com/gin-contrib/multitemplate"
)

type (
	// Mailer allows sending email via SMTP protocol.
	Mailer struct {
		htmlRenderer multitemplate.Renderer
	}
)

// NewMailer initializes Mailer instance.
func NewMailer(c *appysupport.Config, l *appysupport.Logger, s *appyhttp.Server) *Mailer {
	return &Mailer{
		htmlRenderer: s.HTMLRenderer(),
	}
}

// HTML returns the mailer HTML content.
func (m Mailer) HTML(name string, data interface{}) ([]byte, error) {
	recorder := httptest.NewRecorder()
	renderer := m.htmlRenderer.Instance(name, data)
	if err := renderer.Render(recorder); err != nil {
		return nil, err
	}

	return recorder.Body.Bytes(), nil
}

// Text returns the mailer text content.
func (m Mailer) Text(name string, data interface{}) ([]byte, error) {
	return nil, nil
}
