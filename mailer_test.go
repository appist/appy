package appy_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy"
)

type MailerSuite struct {
	appy.TestSuite
	asset       *appy.Asset
	config      *appy.Config
	i18n        *appy.I18n
	logger      *appy.Logger
	previewMail appy.Mail
	server      *appy.Server
	support     appy.Supporter
}

func (s *MailerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &appy.Support{}
	s.logger, _, _ = appy.NewFakeLogger()
	s.asset = appy.NewAsset(http.Dir("testdata/mailer"), map[string]string{
		"docker": "testdata/mailer/.docker",
		"config": "testdata/mailer/configs",
		"locale": "testdata/mailer/pkg/locales",
		"view":   "testdata/mailer/pkg/views",
		"web":    "testdata/mailer/web",
	}, "")
	s.config = appy.NewConfig(s.asset, s.logger, s.support)
	s.i18n = appy.NewI18n(s.asset, s.config, s.logger)
	s.server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	s.server.Use(appy.Recovery(s.logger))
	s.previewMail = appy.Mail{
		From:    "support@appist.io",
		To:      []string{"jane@appist.io"},
		ReplyTo: []string{"john@appist.io", "mary@appist.io"},
		Cc:      []string{"elaine@appist.io", "kerry@appist.io"},
		Bcc:     []string{"joel@appist.io", "daniel@appist.io"},
	}
}

func (s *MailerSuite) TestNewMailerWithDebugBuild() {
	mailer := appy.NewMailer(s.asset, s.config, s.i18n, s.logger, s.server, nil)
	mailer.SetupPreview()

	mail := s.previewMail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = appy.H{
		"username": "cayter",
	}
	mail.Attachments = []string{"testdata/mailer/attachments/fake.txt"}
	mailer.AddPreview(mail)
	s.Error(mailer.Deliver(mail))
	s.Equal(0, len(mailer.Deliveries()))

	w := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath, nil, nil)
	s.Equal(1, len(mailer.Previews()))
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), `data-name="mailers/user/verify_account">mailers/user/verify_account</a>`)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=html&name=mailers/user/verify_account", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "<div>\n  Welcome cayter!\n  测试\n  Hi, John Doe! You have 2 messages.\n</div>")

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=txt&name=mailers/user/verify_account", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "Welcome cayter! 测试 Hi, John Doe! You have 2 messages.")

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=html&name=mailers/user/missing", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)

	mail = s.previewMail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/html_error"
	mail.TemplateData = appy.H{
		"username": "cayter",
	}
	mailer.AddPreview(mail)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=html&name=mailers/user/html_error", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)

	mail = s.previewMail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/txt_error"
	mail.TemplateData = appy.H{
		"username": "cayter",
	}
	mailer.AddPreview(mail)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=txt&name=mailers/user/txt_error", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)

	mail = s.previewMail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = appy.H{
		"username": "cayter",
	}
	mail.Attachments = []string{"testdata/mailer/attachments/missing.txt"}
	mailer.AddPreview(mail)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=html&name=mailers/user/verify_account", nil, nil)
	s.Equal(http.StatusInternalServerError, w.Code)
}

func (s *MailerSuite) TestNewMailerWithReleaseBuild() {
	appy.Build = appy.ReleaseBuild
	defer func() {
		appy.Build = appy.DebugBuild
	}()

	mailer := appy.NewMailer(s.asset, s.config, s.i18n, s.logger, s.server, nil)
	mailer.SetupPreview()

	mail := s.previewMail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = appy.H{
		"username": "cayter",
	}
	mailer.AddPreview(mail)
	s.Error(mailer.Deliver(mail))
	s.Equal(0, len(mailer.Deliveries()))

	w := s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath, nil, nil)
	s.Equal(1, len(mailer.Previews()))
	s.Equal(http.StatusNotFound, w.Code)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=html&name=mailers/user/verify_account", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)

	w = s.server.TestHTTPRequest("GET", s.config.MailerPreviewPath+"/preview?locale=en&ext=txt&name=mailers/user/verify_account", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *MailerSuite) TestMailerWithTestAppyEnv() {
	s.config.AppyEnv = "test"
	mailer := appy.NewMailer(s.asset, s.config, s.i18n, s.logger, s.server, nil)
	mail := s.previewMail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = appy.H{
		"username": "cayter",
	}
	mailer.AddPreview(mail)
	s.NoError(mailer.Deliver(mail))
	s.Equal(1, len(mailer.Deliveries()))
}

func TestMailerSuite(t *testing.T) {
	appy.RunTestSuite(t, new(MailerSuite))
}
