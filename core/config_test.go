package core

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/appist/appy/test"
)

type ConfigSuite struct {
	test.Suite
	logger        *AppLogger
	oldConfigPath string
}

func (s *ConfigSuite) SetupTest() {
	Build = "debug"
	s.logger, _ = newLogger(newLoggerConfig())
	s.oldConfigPath = SSRPaths["config"]
}

func (s *ConfigSuite) TearDownTest() {
	SSRPaths["config"] = s.oldConfigPath
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigDefaultValue() {
	tests := map[string]interface{}{
		"AppyEnv":                         "development",
		"HTTPDebugEnabled":                false,
		"HTTPLogFilterParameters":         []string{"password"},
		"HTTPHealthCheckURL":              "/health_check",
		"HTTPHost":                        "localhost",
		"HTTPPort":                        "3000",
		"HTTPGracefulTimeout":             30 * time.Second,
		"HTTPIdleTimeout":                 75 * time.Second,
		"HTTPMaxHeaderBytes":              0,
		"HTTPReadTimeout":                 60 * time.Second,
		"HTTPReadHeaderTimeout":           60 * time.Second,
		"HTTPWriteTimeout":                60 * time.Second,
		"HTTPSSLEnabled":                  false,
		"HTTPSSLCertPath":                 "./tmp/ssl",
		"HTTPSessionCookieDomain":         "localhost",
		"HTTPSessionCookieHTTPOnly":       true,
		"HTTPSessionCookieMaxAge":         0,
		"HTTPSessionCookiePath":           "/",
		"HTTPSessionCookieSecure":         true,
		"HTTPSessionRedisAddr":            "localhost:6379",
		"HTTPSessionRedisAuth":            "",
		"HTTPSessionRedisDb":              "0",
		"HTTPSessionRedisMaxActive":       0,
		"HTTPSessionRedisMaxIdle":         32,
		"HTTPSessionRedisIdleTimeout":     30 * time.Second,
		"HTTPSessionRedisMaxConnLifetime": 0 * time.Second,
		"HTTPSessionRedisWait":            true,
		"HTTPSessionName":                 "_session",
		"HTTPSessionProvider":             "cookie",
		"HTTPSessionSecrets":              [][]byte{},
		"HTTPAllowedHosts":                []string{},
		"HTTPCSRFCookieDomain":            "localhost",
		"HTTPCSRFCookieHTTPOnly":          true,
		"HTTPCSRFCookieMaxAge":            43200,
		"HTTPCSRFCookieName":              "_csrf_token",
		"HTTPCSRFCookiePath":              "/",
		"HTTPCSRFCookieSecure":            true,
		"HTTPCSRFFieldName":               "authenticity_token",
		"HTTPCSRFRequestHeader":           "X-CSRF-Token",
		"HTTPCSRFSecret":                  []byte{},
		"HTTPSSLRedirect":                 false,
		"HTTPSSLTemporaryRedirect":        false,
		"HTTPSSLHost":                     "",
		"HTTPSTSSeconds":                  int64(0),
		"HTTPSTSIncludeSubdomains":        false,
		"HTTPFrameDeny":                   false,
		"HTTPCustomFrameOptionsValue":     "",
		"HTTPContentTypeNosniff":          false,
		"HTTPBrowserXSSFilter":            false,
		"HTTPContentSecurityPolicy":       "",
		"HTTPReferrerPolicy":              "",
		"HTTPIENoOpen":                    false,
		"HTTPSSLProxyHeaders":             map[string]string{},
	}

	c, _, _ := newConfig(nil, nil, s.logger)
	cv := reflect.ValueOf(c)
	for key, defaultVal := range tests {
		fv := cv.FieldByName(key)

		// An exception case to handle a different host in test for Github actions.
		if key == "HTTPSessionRedisAddr" && os.Getenv("HTTP_SESSION_REDIS_ADDR") != "" {
			s.Equal(fv.Interface(), os.Getenv("HTTP_SESSION_REDIS_ADDR"))
			continue
		}

		switch fv.Kind() {
		case reflect.Map:
			switch fv.Type().String() {
			case "map[string]string":
				for key, val := range fv.Interface().(map[string]string) {
					s.Equal(val, defaultVal.(map[string]string)[key])
				}
			}
		case reflect.Slice, reflect.Array:
			switch fv.Type().String() {
			case "[]string":
				s.Equal(len(fv.Interface().([]string)), len(defaultVal.([]string)))

				for idx, val := range fv.Interface().([]string) {
					s.Equal(val, defaultVal.([]string)[idx])
				}
			case "[]uint8":
				s.Equal(len(fv.Interface().([]uint8)), len(defaultVal.([]uint8)))

				for idx, val := range fv.Interface().([]uint8) {
					s.Equal(val, defaultVal.([]uint8)[idx])
				}
			case "[][]uint8":
				s.Equal(len(fv.Interface().([][]uint8)), len(defaultVal.([][]uint8)))

				for idx, val := range fv.Interface().([][]uint8) {
					s.Equal(val, defaultVal.([][]uint8)[idx])
				}
			default:
				s.Equal(fv.Interface(), defaultVal)
			}
		default:
			if fv.Interface() != defaultVal {
				fmt.Println(key)
			}
			s.Equal(fv.Interface(), defaultVal)
		}
	}
}

func (s *ConfigSuite) TestNewConfigRequiredConfig() {
	os.Setenv("APPY_ENV", "invalid")
	Build = "release"
	c, _, _ := newConfig(http.Dir("./testdata/config"), nil, s.logger)
	s.Equal(false, c.HTTPSSLEnabled)
}

func (s *ConfigSuite) TestNewConfigWithUnparsableEnvVariable() {
	os.Setenv("HTTP_DEBUG_ENABLED", "nil")
	_, _, _ = newConfig(nil, nil, s.logger)
	os.Unsetenv("HTTP_DEBUG_ENABLED")
}

func (s *ConfigSuite) TestMasterKeyWithMissingKeyFile() {
	_, err := MasterKey()
	s.EqualError(err, "open app/config/development.key: no such file or directory")
}

func (s *ConfigSuite) TestMasterKeyWithAppyEnv() {
	os.Setenv("APPY_ENV", "staging")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	key, err := MasterKey()
	s.NoError(err)
	s.Equal([]byte("dummy"), key)
}

func (s *ConfigSuite) TestMasterKeyWithAppyMasterKey() {
	os.Setenv("APPY_MASTER_KEY", "dummy")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	key, err := MasterKey()
	s.NoError(err)
	s.Equal([]byte("dummy"), key)
}

func (s *ConfigSuite) TestMasterKeyWithZeroLength() {
	os.Setenv("APPY_ENV", "empty")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	_, err := MasterKey()
	s.EqualError(err, "the master key should not be blank")
}

func (s *ConfigSuite) TestUndecryptableConfigFallbackToDefault() {
	os.Setenv("APPY_ENV", "undecryptable")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	c, _, _ := newConfig(nil, nil, s.logger)
	s.Equal("3000", c.HTTPPort)
}

func (s *ConfigSuite) TestNonHTTPServeCommand() {
	type config struct {
		AppName string `env:"APP_NAME" envDefault:"tester"`
	}

	os.Args = append(os.Args, "build")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	appConf := &config{}
	appyConf, _, _ := newConfig(nil, appConf, s.logger)
	s.Equal("3000", appyConf.HTTPPort)
	s.Equal("tester", appConf.AppName)
	os.Args = os.Args[:len(os.Args)-1]
}

func (s *ConfigSuite) TestHTTPServeCommand() {
	type config struct {
		AppName string `env:"APP_NAME" envDefault:"tester"`
		AppType string `env:"APP_TYPE,required" envDefault:"b2b"`
	}

	os.Args = append(os.Args, "serve")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	appConf := &config{}
	appyConf, _, _ := newConfig(nil, appConf, s.logger)
	s.Equal("3000", appyConf.HTTPPort)
	s.Equal("tester", appConf.AppName)
	s.Equal("", appConf.AppType)
	os.Args = os.Args[:len(os.Args)-1]
}

func (s *ConfigSuite) TestHTTPServeCommandDebugBuildWithUndecryptableConfig() {
	type config struct {
		AppName string `env:"APP_NAME" envDefault:"tester"`
	}

	os.Setenv("APPY_ENV", "undecryptable")
	os.Args = append(os.Args, "serve")
	SSRPaths["config"] = "./testdata/.ssr/app/config"
	appConf := &config{}
	appyConf, _, _ := newConfig(nil, appConf, s.logger)
	s.Equal("3000", appyConf.HTTPPort)
	s.Equal("tester", appConf.AppName)
	os.Args = os.Args[:len(os.Args)-1]
}

func (s *ConfigSuite) TestHTTPServeCommandReleaseBuildWithUndecryptableConfig() {
	type config struct {
		AppName string `env:"APP_NAME" envDefault:"tester"`
		AppType string `env:"APP_TYPE,required" envDefault:"b2b"`
	}

	Build = "release"
	os.Setenv("APPY_ENV", "undecryptable")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Args = append(os.Args, "serve")
	appConf := &config{}
	appyConf, _, _ := newConfig(http.Dir("./testdata"), appConf, s.logger)
	s.Equal("3000", appyConf.HTTPPort)
	s.Equal("tester", appConf.AppName)
	s.Equal("", appConf.AppType)
	os.Args = os.Args[:len(os.Args)-1]
}

func TestConfig(t *testing.T) {
	test.Run(t, new(ConfigSuite))
}
