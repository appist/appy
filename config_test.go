package appy_test

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/appist/appy"
)

type ConfigSuite struct {
	appy.TestSuite
	asset   *appy.Asset
	logger  *appy.Logger
	support appy.Supporter
}

func (s *ConfigSuite) SetupTest() {
	layout := map[string]string{
		"docker": "testdata/config/.docker",
		"config": "testdata/config/configs",
		"locale": "testdata/config/pkg/locales",
		"view":   "testdata/config/pkg/views",
		"web":    "testdata/config/web",
	}
	s.asset = appy.NewAsset(nil, layout, "")
	s.logger, _, _ = appy.NewFakeLogger()
	s.support = &appy.Support{}
}

func (s *ConfigSuite) TestNewConfig() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	tt := map[string]interface{}{
		"AppyEnv":                         "development",
		"AssetHost":                       "",
		"GQLPlaygroundEnabled":            false,
		"GQLPlaygroundPath":               "/docs/graphql",
		"GQLAPQCacheSize":                 100,
		"GQLQueryCacheSize":               1000,
		"GQLComplexityLimit":              1000,
		"GQLMultipartMaxMemory":           int64(0),
		"GQLMultipartMaxUploadSize":       int64(0),
		"GQLWebsocketKeepAliveDuration":   10 * time.Second,
		"HTTPGzipCompressLevel":           -1,
		"HTTPGzipExcludedExts":            []string{},
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
		"HTTPSessionRedisAddr":            "localhost:6379",
		"HTTPSessionRedisAuth":            "",
		"HTTPSessionRedisDB":              "0",
		"HTTPSessionRedisMaxActive":       64,
		"HTTPSessionRedisMaxIdle":         32,
		"HTTPSessionRedisIdleTimeout":     30 * time.Second,
		"HTTPSessionRedisMaxConnLifetime": 30 * time.Second,
		"HTTPSessionRedisWait":            true,
		"HTTPSessionCookieName":           "_session",
		"HTTPSessionProvider":             "cookie",
		"HTTPSessionExpiration":           1209600,
		"HTTPSessionSecrets":              [][]byte{},
		"HTTPSessionCookieDomain":         "localhost",
		"HTTPSessionCookieHTTPOnly":       true,
		"HTTPSessionCookiePath":           "/",
		"HTTPSessionCookieSameSite":       http.SameSite(1),
		"HTTPSessionCookieSecure":         false,
		"HTTPAllowedHosts":                []string{},
		"HTTPCSRFCookieDomain":            "localhost",
		"HTTPCSRFCookieHTTPOnly":          true,
		"HTTPCSRFCookieMaxAge":            0,
		"HTTPCSRFCookieName":              "_csrf_token",
		"HTTPCSRFCookiePath":              "/",
		"HTTPCSRFCookieSameSite":          http.SameSite(1),
		"HTTPCSRFCookieSecure":            false,
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
		"HTTPSSLProxyHeaders":             map[string]string{"X-Forwarded-Proto": "https"},
		"I18nDefaultLocale":               "en",
		"MailerSMTPAddr":                  "",
		"MailerSMTPPlainAuthIdentity":     "",
		"MailerSMTPPlainAuthUsername":     "",
		"MailerSMTPPlainAuthPassword":     "",
		"MailerSMTPPlainAuthHost":         "",
		"MailerPreviewBaseURL":            "/appy/mailers",
		"WorkerRedisSentinelAddrs":        []string{},
		"WorkerRedisSentinelDB":           0,
		"WorkerRedisSentinelMasterName":   "",
		"WorkerRedisSentinelPassword":     "",
		"WorkerRedisSentinelPoolSize":     25,
		"WorkerRedisAddr":                 "",
		"WorkerRedisDB":                   0,
		"WorkerRedisPassword":             "",
		"WorkerRedisPoolSize":             25,
		"WorkerRedisURL":                  "",
		"WorkerConcurrency":               25,
		"WorkerQueues":                    "default:10",
		"WorkerStrictPriority":            false,
	}

	config := appy.NewConfig(s.asset, s.logger, s.support)
	cv := reflect.ValueOf(*config)
	for key, defaultVal := range tt {
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

	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config = appy.NewConfig(s.asset, s.logger, s.support)
	s.Equal([]byte("481e5d98a31585148b8b1dfb6a3c0465"), config.MasterKey())
	s.Nil(config.Errors())
	s.Equal("testdata/config/configs/.env.development", config.Path())
}

func (s *ConfigSuite) TestNewConfigMissingRequiredConfig() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.NotNil(config.Errors())
	s.EqualError(config.Errors()[0], `required environment variable "HTTP_SESSION_SECRETS" is not set. required environment variable "HTTP_CSRF_SECRET" is not set`)
}

func (s *ConfigSuite) TestNewConfigWithEnvVariableOverride() {
	os.Setenv("APPY_ENV", "override")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_PORT", "5000")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_PORT")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.Equal(config.HTTPPort, "5000")
}

func (s *ConfigSuite) TestNewConfigWithMissingMasterKey() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.EqualError(config.Errors()[0], appy.ErrReadMasterKeyFile.Error())
}

func (s *ConfigSuite) TestNewConfigWithEmptyMasterKeyFile() {
	os.Setenv("APPY_ENV", "empty")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.EqualError(config.Errors()[0], appy.ErrMissingMasterKey.Error())
}

func (s *ConfigSuite) TestNewConfigWithUnparsableEnvVariable() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("GQL_PLAYGROUND_ENABLED", "nil")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("GQL_PLAYGROUND_ENABLED")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.Contains(config.Errors()[0].Error(), `strconv.ParseBool: parsing "nil": invalid syntax.`)
}

func (s *ConfigSuite) TestNewConfigWithUndecryptableConfig() {
	os.Setenv("APPY_ENV", "undecryptable")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.Contains(config.Errors()[0].Error(), "unable to decrypt 'HTTP_PORT' value in 'testdata/config/configs/.env.undecryptable'")
}

func (s *ConfigSuite) TestNewConfigWithUnparsableConfig() {
	os.Setenv("APPY_ENV", "unparsable")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.Contains(config.Errors()[0].Error(), "Can't separate key from value")
}

func (s *ConfigSuite) TestNewConfigWithInvalidAssetsPath() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	config := appy.NewConfig(appy.NewAsset(nil, nil, ""), s.logger, s.support)
	s.Contains(config.Errors()[0].Error(), "open configs/.env.development: no such file or directory")
}

func (s *ConfigSuite) TestIsProtectedEnv() {
	config := appy.NewConfig(s.asset, s.logger, s.support)
	s.Equal(false, config.IsProtectedEnv())

	os.Setenv("APPY_ENV", "production")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config = appy.NewConfig(s.asset, s.logger, s.support)
	s.Equal(true, config.IsProtectedEnv())
}

func TestConfigSuite(t *testing.T) {
	appy.RunTestSuite(t, new(ConfigSuite))
}
