package support

import (
	"net/http"
	"os"
	"testing"

	"github.com/appist/appy/test"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type i18nSuite struct {
	test.Suite
	asset  *Asset
	config *Config
	logger *Logger
}

func (s *i18nSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = NewTestLogger()
}

func (s *i18nSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *i18nSuite) TestMissingLocales() {
	s.asset = NewAsset(nil, "testdata/missing")
	s.config = NewConfig(s.asset, s.logger)

	s.Panics(func() { NewI18n(s.asset, s.config, s.logger) })
}

func (s *i18nSuite) TestTWithDebugBuild() {
	s.asset = NewAsset(nil, "testdata/i18n/t_with_debug_build")
	s.config = NewConfig(s.asset, s.logger)
	i18n := NewI18n(s.asset, s.config, s.logger)

	s.NotNil(i18n.Bundle())
	s.ElementsMatch([]string{"en", "zh-TW", "zh-CN"}, i18n.Locales())
	s.Equal("", i18n.T("title.foo", "en"))

	s.Equal("Test", i18n.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", i18n.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", i18n.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", i18n.T("body.message", 2, H{"Name": "tester"}))

	s.Equal("測試", i18n.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", i18n.T("body.message", 0, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", i18n.T("body.message", 1, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", i18n.T("body.message", 2, H{"Name": "tester"}, "zh-TW"))
}

func (s *i18nSuite) TestTWithReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	s.asset = NewAsset(http.Dir("testdata/i18n/t_with_release_build"), "")
	s.config = NewConfig(s.asset, s.logger)
	i18n := NewI18n(s.asset, s.config, s.logger)

	s.NotNil(i18n.Bundle())
	s.ElementsMatch([]string{"en", "zh-TW", "zh-CN"}, i18n.Locales())
	s.Equal("", i18n.T("title.foo", "en"))

	s.Equal("Test", i18n.T("title.test"))
	s.Equal("Hi, tester! You have 0 message.", i18n.T("body.message", 0, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 1 message.", i18n.T("body.message", 1, H{"Name": "tester"}))
	s.Equal("Hi, tester! You have 2 messages.", i18n.T("body.message", 2, H{"Name": "tester"}))

	s.Equal("測試", i18n.T("title.test", "zh-TW"))
	s.Equal("嗨, tester! 您有0則訊息。", i18n.T("body.message", 0, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有1則訊息。", i18n.T("body.message", 1, H{"Name": "tester"}, "zh-TW"))
	s.Equal("嗨, tester! 您有2則訊息。", i18n.T("body.message", 2, H{"Name": "tester"}, "zh-TW"))
}

func (s *i18nSuite) TestValidationErrors() {
	s.asset = NewAsset(nil, "../record/testdata")
	s.config = NewConfig(s.asset, s.logger)
	i18n := NewI18n(s.asset, s.config, s.logger)
	validator, _ := binding.Validator.Engine().(*validator.Validate)

	{
		type user1 struct {
			Email ZString `db:"email" binding:"required"`
		}

		user := user1{}

		errs := i18n.ValidationErrors(validator.Struct(user), "")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "user1.Email must not be blank")

		errs = i18n.ValidationErrors(validator.Struct(user), "zh-CN")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "user1.Email must not be blank")

		errs = i18n.ValidationErrors(validator.Struct(user), "zh-TW")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "user1.Email must not be blank")
	}

	{
		type user2 struct {
			Password             string  `db:"password"`
			PasswordConfirmation NString `db:"password_confirmation" binding:"eqfield=Password"`
		}

		user := user2{Password: "foo", PasswordConfirmation: NewNString("foobar")}

		errs := i18n.ValidationErrors(validator.Struct(user), "")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "password confirmation (foobar) must be equal to password")

		errs = i18n.ValidationErrors(validator.Struct(user), "zh-CN")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "确认密码(foobar)必须与密码相同")

		errs = i18n.ValidationErrors(validator.Struct(user), "zh-TW")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "確認密碼(foobar)必須與密碼相同")
	}

	{
		type user3 struct {
			Username string `db:"age" binding:"min=5,max=8"`
		}

		user := user3{Username: "foo"}

		errs := i18n.ValidationErrors(validator.Struct(user), "")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "user3.Username cannot be less than 5")

		errs = i18n.ValidationErrors(validator.Struct(user), "zh-CN")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "user3.Username不能小于5")

		errs = i18n.ValidationErrors(validator.Struct(user), "zh-TW")
		s.Equal(1, len(errs))
		s.EqualError(errs[0], "user3.Username不能小於5")
	}
}

func TestI18nSuite(t *testing.T) {
	test.Run(t, new(i18nSuite))
}
