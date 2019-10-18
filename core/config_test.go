package core

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/appist/appy/test"
)

type ConfigSuite struct {
	test.Suite
}

func (s *ConfigSuite) SetupTest() {
	Build = "debug"
}

func (s *ConfigSuite) TearDownTest() {
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

	c, _ := newConfig(nil)
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
			s.Equal(fv.Interface(), defaultVal)
		}
	}
}

func (s *ConfigSuite) TestNewConfigRequiredConfig() {
	os.Setenv("APPY_ENV", "invalid")
	Build = "release"
	c, err := newConfig(http.Dir("./testdata/config"))
	s.EqualError(err, "required environment variable \"HTTP_SESSION_SECRETS\" is not set. required environment variable \"HTTP_CSRF_SECRET\" is not set")
	s.Equal(false, c.HTTPSSLEnabled)
	os.Unsetenv("APPY_ENV")
}

func (s *ConfigSuite) TestNewConfigWithReleaseBuild() {
	Build = "release"
	c, err := newConfig(http.Dir("./testdata"))
	s.Nil(err)
	s.Equal(true, c.HTTPSSLEnabled)
}

func (s *ConfigSuite) TestNewConfigWithUnparsableEnvVariable() {
	os.Setenv("HTTP_DEBUG_ENABLED", "nil")
	_, err := newConfig(nil)
	os.Unsetenv("HTTP_DEBUG_ENABLED")
	s.NotNil(err)
}

func (s *ConfigSuite) TestMasterKeyWithMissingKeyFile() {
	_, err := MasterKey()
	s.EqualError(err, "open config/development.key: no such file or directory")
}

func (s *ConfigSuite) TestMasterKeyWithAppyEnv() {
	oldSSRConfig := SSRPaths["config"]
	SSRPaths["config"] = "./testdata/.ssr/config"
	os.Setenv("APPY_ENV", "staging")
	key, err := MasterKey()
	s.NoError(err)
	s.Equal([]byte("dummy"), key)
	os.Unsetenv("APPY_ENV")
	SSRPaths["config"] = oldSSRConfig
}

func (s *ConfigSuite) TestMasterKeyWithAppyMasterKey() {
	oldSSRConfig := SSRPaths["config"]
	SSRPaths["config"] = "./testdata/.ssr/config"
	os.Setenv("APPY_MASTER_KEY", "dummy")
	key, err := MasterKey()
	s.NoError(err)
	s.Equal([]byte("dummy"), key)
	os.Unsetenv("APPY_MASTER_KEY")
	SSRPaths["config"] = oldSSRConfig
}

func (s *ConfigSuite) TestMasterKeyWithZeroLength() {
	oldSSRConfig := SSRPaths["config"]
	SSRPaths["config"] = "./testdata/.ssr/config"
	os.Setenv("APPY_ENV", "empty")
	_, err := MasterKey()
	s.EqualError(err, "the master key cannot be blank, please either pass in \"APPY_MASTER_KEY\" environment variable or store it in \"config/<APPY_ENV>.key\"")
	os.Unsetenv("APPY_ENV")
	SSRPaths["config"] = oldSSRConfig
}

func TestConfig(t *testing.T) {
	test.Run(t, new(ConfigSuite))
}
