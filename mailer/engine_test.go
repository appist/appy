package mailer

import (
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type mailerSuite struct {
	test.Suite
	asset  *support.Asset
	config *support.Config
	i18n   *support.I18n
	logger *support.Logger
	mail   *Mail
}

func (s *mailerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(http.Dir("testdata/default"), "testdata/default")
	s.config = support.NewConfig(s.asset, s.logger)
	s.i18n = support.NewI18n(s.asset, s.config, s.logger)
	s.mail = &Mail{
		From:    "support@appist.io",
		To:      []string{"jane@appist.io"},
		ReplyTo: []string{"john@appist.io", "mary@appist.io"},
		Cc:      []string{"elaine@appist.io", "kerry@appist.io"},
		Bcc:     []string{"joel@appist.io", "daniel@appist.io"},
	}
}

func (s *mailerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mailerSuite) TestNewEngineWithDebugBuild() {
	mail := s.mail
	mail.Subject = "mailers.user.verifyAccount.subject"

	{
		mail.Template = "mailers/user/missing"

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Equal("template mailers/user/missing.html can't be loaded", err.Error())
	}

	{
		mail.Template = "mailers/user/error"

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Contains(err.Error(), "Jet Runtime Error(\"mailers/user/error.html\":0): there is no field or method \"foobar\" in")
	}

	{
		mail.Template = "mailers/user/reset_password"

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Equal("template mailers/user/reset_password.txt can't be loaded", err.Error())
	}

	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = support.H{
		"username": "cayter",
	}

	{
		mail.Attachments = []string{"testdata/missing/attachments/fake.txt"}

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Equal("open testdata/missing/attachments/fake.txt: no such file or directory", err.Error())
	}

	{
		mail.Attachments = []string{"testdata/default/attachments/fake.txt"}

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		mailer.AddPreview(mail)

		previews := mailer.Previews()
		s.Equal(1, len(previews))

		err := mailer.Deliver(mail)
		s.Equal("dial tcp: missing address", err.Error())

		deliveries := mailer.Deliveries()
		s.Equal(0, len(deliveries))
	}
}

func (s *mailerSuite) TestNewEngineWithReleaseBuild() {
	support.Build = support.ReleaseBuild
	defer func() {
		support.Build = support.DebugBuild
	}()

	mail := s.mail
	mail.Subject = "mailers.user.verifyAccount.subject"

	{
		mail.Template = "mailers/user/missing"

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Equal("template mailers/user/missing.html can't be loaded", err.Error())
	}

	{
		mail.Template = "mailers/user/error"

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Contains(err.Error(), "Jet Runtime Error(\"mailers/user/error.html\":0): there is no field or method \"foobar\" in")
	}

	{
		mail.Template = "mailers/user/reset_password"

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Equal("template mailers/user/reset_password.txt can't be loaded", err.Error())
	}

	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = support.H{
		"username": "cayter",
	}

	{
		mail.Attachments = []string{"testdata/missing/attachments/fake.txt"}
		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		err := mailer.Deliver(mail)

		s.Equal("open testdata/missing/attachments/fake.txt: no such file or directory", err.Error())
	}

	{
		mail.Attachments = []string{"testdata/default/attachments/fake.txt"}

		mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
		mailer.AddPreview(mail)

		previews := mailer.Previews()
		s.Equal(1, len(previews))

		err := mailer.Deliver(mail)
		s.Equal("dial tcp: missing address", err.Error())

		deliveries := mailer.Deliveries()
		s.Equal(0, len(deliveries))
	}
}

func (s *mailerSuite) TestMailerWithTestAppyEnv() {
	s.config.AppyEnv = "test"

	mail := s.mail
	mail.Subject = "mailers.user.verifyAccount.subject"
	mail.Template = "mailers/user/verify_account"
	mail.TemplateData = support.H{
		"username": "cayter",
	}

	mailer := NewEngine(s.asset, s.config, s.i18n, s.logger, nil)
	mailer.AddPreview(mail)

	previews := mailer.Previews()
	s.Equal(1, len(previews))

	err := mailer.Deliver(mail)
	s.Nil(err)

	deliveries := mailer.Deliveries()
	s.Equal(1, len(deliveries))
	s.Contains(deliveries[0].HTML, "cayter")
	s.Contains(deliveries[0].HTML, "测试")
	s.Contains(deliveries[0].HTML, "Hi, John Doe! You have 2 messages.")
	s.Contains(deliveries[0].Text, "cayter")
	s.Contains(deliveries[0].Text, "测试")
	s.Contains(deliveries[0].Text, "Hi, John Doe! You have 2 messages.")
}

func TestMailerSuite(t *testing.T) {
	test.Run(t, new(mailerSuite))
}
