package support

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/appist/appy/test"
)

type configSuite struct {
	test.Suite
	logger *Logger
}

func (s *configSuite) SetupTest() {
	s.logger, _, _ = NewTestLogger()
}

func (s *configSuite) TestDefaultValue() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	tt := map[string]interface{}{
		"AppyEnv":                            "development",
		"AssetHost":                          "",
		"GQLPlaygroundEnabled":               false,
		"GQLPlaygroundPath":                  "/docs/graphql",
		"GQLAPQCacheSize":                    100,
		"GQLQueryCacheSize":                  1000,
		"GQLComplexityLimit":                 1000,
		"GQLMultipartMaxMemory":              int64(0),
		"GQLMultipartMaxUploadSize":          int64(0),
		"GQLWebsocketKeepAliveDuration":      10 * time.Second,
		"HTTPGzipCompressLevel":              -1,
		"HTTPGzipExcludedExts":               []string{},
		"HTTPLogFilterParameters":            []string{"password"},
		"HTTPHealthCheckPath":                "/health_check",
		"HTTPHost":                           "localhost",
		"HTTPPort":                           "3000",
		"HTTPGracefulShutdownTimeout":        30 * time.Second,
		"HTTPIdleTimeout":                    75 * time.Second,
		"HTTPMaxHeaderBytes":                 0,
		"HTTPReadTimeout":                    60 * time.Second,
		"HTTPReadHeaderTimeout":              60 * time.Second,
		"HTTPWriteTimeout":                   60 * time.Second,
		"HTTPSSLEnabled":                     false,
		"HTTPSSLCertPath":                    "./tmp/ssl",
		"HTTPSessionRedisAddr":               "localhost:6379",
		"HTTPSessionRedisPassword":           "",
		"HTTPSessionRedisDB":                 0,
		"HTTPSessionRedisMaxConnAge":         0 * time.Second,
		"HTTPSessionRedisMinIdleConns":       0,
		"HTTPSessionRedisIdleCheckFrequency": 1 * time.Minute,
		"HTTPSessionRedisIdleTimeout":        5 * time.Minute,
		"HTTPSessionRedisPoolSize":           10,
		"HTTPSessionRedisPoolTimeout":        4 * time.Second,
		"HTTPSessionCookieName":              "_session",
		"HTTPSessionProvider":                "cookie",
		"HTTPSessionExpiration":              1209600,
		"HTTPSessionSecrets":                 [][]byte{},
		"HTTPSessionCookieDomain":            "localhost",
		"HTTPSessionCookieHTTPOnly":          true,
		"HTTPSessionCookiePath":              "/",
		"HTTPSessionCookieSameSite":          http.SameSite(1),
		"HTTPSessionCookieSecure":            false,
		"HTTPAllowedHosts":                   []string{},
		"HTTPCSRFCookieDomain":               "localhost",
		"HTTPCSRFCookieHTTPOnly":             true,
		"HTTPCSRFCookieMaxAge":               0,
		"HTTPCSRFCookieName":                 "_csrf_token",
		"HTTPCSRFCookiePath":                 "/",
		"HTTPCSRFCookieSameSite":             http.SameSite(1),
		"HTTPCSRFCookieSecure":               false,
		"HTTPCSRFAuthenticityFieldName":      "authenticity_token",
		"HTTPCSRFRequestHeader":              "X-CSRF-Token",
		"HTTPCSRFSecret":                     []byte{},
		"HTTPSSLRedirect":                    false,
		"HTTPSSLTemporaryRedirect":           false,
		"HTTPSSLHost":                        "",
		"HTTPSTSSeconds":                     int64(0),
		"HTTPSTSIncludeSubdomains":           false,
		"HTTPFrameDeny":                      false,
		"HTTPCustomFrameOptionsValue":        "",
		"HTTPContentTypeNosniff":             false,
		"HTTPBrowserXSSFilter":               false,
		"HTTPContentSecurityPolicy":          "",
		"HTTPReferrerPolicy":                 "",
		"HTTPIENoOpen":                       false,
		"HTTPSSLProxyHeaders":                map[string]string{"X-Forwarded-Proto": "https"},
		"I18nDefaultLocale":                  "en",
		"MailerSMTPAddr":                     "",
		"MailerSMTPPlainAuthIdentity":        "",
		"MailerSMTPPlainAuthUsername":        "",
		"MailerSMTPPlainAuthPassword":        "",
		"MailerSMTPPlainAuthHost":            "",
		"MailerPreviewPath":                  "/appy/mailers",
		"WorkerRedisSentinelAddrs":           []string{},
		"WorkerRedisSentinelDB":              0,
		"WorkerRedisSentinelMasterName":      "",
		"WorkerRedisSentinelPassword":        "",
		"WorkerRedisSentinelPoolSize":        25,
		"WorkerRedisAddr":                    "localhost:6379",
		"WorkerRedisDB":                      0,
		"WorkerRedisPassword":                "",
		"WorkerRedisPoolSize":                25,
		"WorkerRedisURL":                     "",
		"WorkerConcurrency":                  25,
		"WorkerQueues":                       "default:10",
		"WorkerStrictPriority":               false,
		"WorkerGracefulShutdownTimeout":      30 * time.Second,
	}

	asset := NewAsset(nil, "")
	config := NewConfig(asset, s.logger)
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
					s.Equal(defaultVal.(map[string]string)[key], val)
				}
			}
		case reflect.Slice, reflect.Array:
			switch fv.Type().String() {
			case "[]string":
				s.Equal(len(defaultVal.([]string)), len(fv.Interface().([]string)))

				for idx, val := range fv.Interface().([]string) {
					s.Equal(defaultVal.([]string)[idx], val)
				}
			case "[]uint8":
				s.Equal(len(defaultVal.([]uint8)), len(fv.Interface().([]uint8)))

				for idx, val := range fv.Interface().([]uint8) {
					s.Equal(defaultVal.([]uint8)[idx], val)
				}
			case "[][]uint8":
				s.Equal(len(defaultVal.([][]uint8)), len(fv.Interface().([][]uint8)))

				for idx, val := range fv.Interface().([][]uint8) {
					s.Equal(defaultVal.([][]uint8)[idx], val)
				}
			default:
				s.Equal(defaultVal, fv.Interface())
			}
		default:
			s.Equal(defaultVal, fv.Interface())
		}
	}
}

func (s *configSuite) TestIsProtectedEnv() {
	{
		asset := NewAsset(nil, "")
		config := NewConfig(asset, s.logger)

		s.Equal(false, config.IsProtectedEnv())
	}

	{
		os.Setenv("APPY_ENV", "production")
		os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
		}()

		asset := NewAsset(nil, "")
		config := NewConfig(asset, s.logger)

		s.Equal(true, config.IsProtectedEnv())
	}
}

func (s *configSuite) TestMasterKeyParsing() {
	{
		asset := NewAsset(nil, "")
		config := NewConfig(asset, s.logger)

		s.Equal(ErrReadMasterKeyFile, config.Errors()[0])
	}

	{
		os.Setenv("APPY_ENV", "development")
		defer func() { os.Unsetenv("APPY_ENV") }()

		asset := NewAsset(nil, "testdata/config/master_key_parsing")
		config := NewConfig(asset, s.logger)

		s.Equal(ErrMissingMasterKey, config.Errors()[0])
	}

	{
		os.Setenv("APPY_ENV", "development")
		os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
		}()

		asset := NewAsset(nil, "")
		config := NewConfig(asset, s.logger)

		s.Equal("58f364f29b568807ab9cffa22c99b538", string(config.MasterKey()))
	}
}

func (s *configSuite) TestRequiredConfigMissing() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	asset := NewAsset(nil, "testdata/config/missing_required_config")
	config := NewConfig(asset, s.logger)

	s.Equal(`required environment variable "HTTP_SESSION_SECRETS" is not set. required environment variable "HTTP_CSRF_SECRET" is not set`, config.Errors()[0].Error())
}

func (s *configSuite) TestConfigFileParsing() {
	{
		os.Setenv("APPY_ENV", "malformed")
		os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
			os.Unsetenv("HTTP_CSRF_SECRET")
			os.Unsetenv("HTTP_SESSION_SECRETS")
		}()

		asset := NewAsset(nil, "testdata/config/config_file_parsing")
		config := NewConfig(asset, s.logger)

		s.Equal("Can't separate key from value", config.Errors()[0].Error())
	}

	{
		os.Setenv("APPY_ENV", "undecryptable")
		os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
			os.Unsetenv("HTTP_CSRF_SECRET")
			os.Unsetenv("HTTP_SESSION_SECRETS")
		}()

		asset := NewAsset(nil, "testdata/config/config_file_parsing")
		config := NewConfig(asset, s.logger)

		s.Equal("unable to decrypt 'GQL_PLAYGROUND_ENABLED' value in 'configs/.env.undecryptable'", config.Errors()[0].Error())
	}

	{
		os.Setenv("APPY_ENV", "undecodable")
		os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
			os.Unsetenv("HTTP_CSRF_SECRET")
			os.Unsetenv("HTTP_SESSION_SECRETS")
		}()

		asset := NewAsset(nil, "testdata/config/config_file_parsing")
		config := NewConfig(asset, s.logger)

		s.Equal("encoding/hex: invalid byte: U+00E6 'æ'", config.Errors()[0].Error())
		os.Unsetenv("HTTP_HOST")
	}

	{
		os.Setenv("APPY_ENV", "decryptable")
		os.Setenv("APPY_MASTER_KEY", "5a9f28ee6301fbaee87d27a9af5cbdc73f3e907f0dec11a4f37e361c1e0687da")
		os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
			os.Unsetenv("HTTP_CSRF_SECRET")
			os.Unsetenv("HTTP_SESSION_SECRETS")
		}()

		asset := NewAsset(nil, "testdata/config/config_file_parsing")
		config := NewConfig(asset, s.logger)

		s.Equal(0, len(config.Errors()))
		s.Equal("0.0.0.0", config.HTTPHost)
	}

	{
		os.Setenv("APPY_ENV", "decryptable")
		os.Setenv("APPY_MASTER_KEY", "5a9f28ee6301fbaee87d27a9af5cbdc73f3e907f0dec11a4f37e361c1e0687da")
		os.Setenv("HTTP_CSRF_SECRET", "58f364f29b568807ab9cffa22c99b538")
		os.Setenv("HTTP_HOST", "1.2.3.4")
		os.Setenv("HTTP_SESSION_SECRETS", "58f364f29b568807ab9cffa22c99b538")
		defer func() {
			os.Unsetenv("APPY_ENV")
			os.Unsetenv("APPY_MASTER_KEY")
			os.Unsetenv("HTTP_CSRF_SECRET")
			os.Unsetenv("HTTP_HOST")
			os.Unsetenv("HTTP_SESSION_SECRETS")
		}()

		asset := NewAsset(nil, "testdata/config/config_file_parsing")
		config := NewConfig(asset, s.logger)

		s.Equal(0, len(config.Errors()))
		s.Equal("1.2.3.4", config.HTTPHost)
	}
}

func TestConfigSuite(t *testing.T) {
	test.Run(t, new(configSuite))
}
