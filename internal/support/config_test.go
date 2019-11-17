package support

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/appist/appy/internal/test"
)

type ConfigSuite struct {
	test.Suite
	oldSSRPaths map[string]string
}

func (s *ConfigSuite) SetupTest() {
	s.oldSSRPaths = SSRPaths
	SSRPaths = map[string]string{
		"root":   "testdata/.ssr",
		"config": "testdata/pkg/config",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
	}
}

func (s *ConfigSuite) TearDownTest() {
	SSRPaths = s.oldSSRPaths
}

func (s *ConfigSuite) TestNewConfigDefaultValue() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer os.Unsetenv("APPY_MASTER_KEY")

	tt := map[string]interface{}{
		"AppyEnv":                         "development",
		"GQLPlaygroundEnabled":            false,
		"GQLPlaygroundPath":               "/docs/graphql",
		"GQLCacheSize":                    1000,
		"GQLComplexityLimit":              200,
		"GQLUploadMaxMemory":              int64(100000000),
		"GQLUploadMaxSize":                int64(100000000),
		"GQLWebsocketKeepAliveDuration":   30 * time.Second,
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
		"HTTPSessionRedisAddr":            "localhost:6379",
		"HTTPSessionRedisAuth":            "",
		"HTTPSessionRedisDb":              "0",
		"HTTPSessionRedisMaxActive":       64,
		"HTTPSessionRedisMaxIdle":         32,
		"HTTPSessionRedisIdleTimeout":     30 * time.Second,
		"HTTPSessionRedisMaxConnLifetime": 30 * time.Second,
		"HTTPSessionRedisWait":            true,
		"HTTPSessionName":                 "_session",
		"HTTPSessionProvider":             "cookie",
		"HTTPSessionSecrets":              [][]byte{},
		"HTTPSessionDomain":               "localhost",
		"HTTPSessionHTTPOnly":             true,
		"HTTPSessionExpiration":           1209600,
		"HTTPSessionPath":                 "/",
		"HTTPSessionSecure":               false,
		"HTTPAllowedHosts":                []string{},
		"HTTPCSRFCookieDomain":            "localhost",
		"HTTPCSRFCookieHTTPOnly":          true,
		"HTTPCSRFCookieMaxAge":            0,
		"HTTPCSRFCookieName":              "_csrf_token",
		"HTTPCSRFCookiePath":              "/",
		"HTTPCSRFCookieSecure":            false,
		"HTTPCSRFFieldName":               "authenticity_token",
		"HTTPCSRFRequestHeader":           "X-CSRF-Token",
		"HTTPCSRFSecret":                  []byte{},
		"HTTPSSLRedirect":                 false,
		"HTTPSSLTemporaryRedirect":        false,
		"HTTPSSLHost":                     "localhost:3443",
		"HTTPSTSSeconds":                  int64(0),
		"HTTPSTSIncludeSubdomains":        false,
		"HTTPFrameDeny":                   true,
		"HTTPCustomFrameOptionsValue":     "",
		"HTTPContentTypeNosniff":          false,
		"HTTPBrowserXSSFilter":            false,
		"HTTPContentSecurityPolicy":       "",
		"HTTPReferrerPolicy":              "",
		"HTTPIENoOpen":                    false,
		"HTTPSSLProxyHeaders":             map[string]string{},
	}

	logger := NewLogger()
	config := NewConfig(logger, nil)
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
}

func (s *ConfigSuite) TestNewConfigWithoutSettingRequiredConfig() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.NotNil(config.Errors())
	s.EqualError(config.Errors()[0], `required environment variable "HTTP_SESSION_SECRETS" is not set. required environment variable "HTTP_CSRF_SECRET" is not set`)
}

func (s *ConfigSuite) TestNewConfigWithSettingRequiredConfig() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.Equal([]byte("481e5d98a31585148b8b1dfb6a3c0465"), config.MasterKey())
	s.Nil(config.Errors())
}

func (s *ConfigSuite) TestNewConfigWithUnparsableEnvVariable() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_DEBUG_ENABLED", "nil")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_DEBUG_ENABLED")
	}()

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.Contains(config.Errors()[0].Error(), `strconv.ParseBool: parsing "nil": invalid syntax.`)
}

func (s *ConfigSuite) TestNewConfigWithUndecryptableConfig() {
	os.Setenv("APPY_ENV", "undecryptable")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.Contains(config.Errors()[0].Error(), "unable to decrypt 'HTTP_PORT' value in 'testdata/pkg/config/.env.undecryptable'")
}

func (s *ConfigSuite) TestNewConfigWithInvalidAssetsPath() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()
	logger := NewLogger()
	config := NewConfig(logger, http.Dir("."))
	s.Contains(config.Errors()[0].Error(), "open testdata/.ssr/testdata/pkg/config/.env.development: no such file or directory")
}

func (s *ConfigSuite) TestNewConfigWithMissingConfigInAssets() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()
	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.EqualError(config.Errors()[0], ErrNoConfigInAssets.Error())
}

func (s *ConfigSuite) TestNewConfigWithUnparsableConfig() {
	os.Setenv("APPY_ENV", "unparsable")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
	}()

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.Contains(config.Errors()[0].Error(), "Can't separate key from value")
}

func (s *ConfigSuite) TestIsConfigErrored() {
	os.Setenv("APPY_ENV", "development")

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.Equal(true, IsConfigErrored(config, logger))

	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	logger = NewLogger()
	config = NewConfig(logger, nil)
	s.Equal(false, IsConfigErrored(config, logger))

	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ConfigSuite) TestIsProtectedEnv() {
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	logger := NewLogger()
	config := NewConfig(logger, nil)
	s.Equal(false, IsProtectedEnv(config))

	os.Setenv("APPY_ENV", "production")
	logger = NewLogger()
	config = NewConfig(logger, nil)
	s.Equal(true, IsProtectedEnv(config))
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ConfigSuite) TestMasterKeyWithMissingKeyFile() {
	_, err := parseMasterKey()
	s.EqualError(err, ErrReadMasterKeyFile.Error())
}

func (s *ConfigSuite) TestMasterKeyWithMissingAppyMasterKey() {
	Build = ReleaseBuild
	_, err := parseMasterKey()
	s.EqualError(err, ErrNoMasterKey.Error())
	Build = DebugBuild
}

func (s *ConfigSuite) TestMasterKeyWithZeroLength() {
	Build = ReleaseBuild
	_, err := parseMasterKey()
	s.EqualError(err, ErrNoMasterKey.Error())
	Build = DebugBuild
}

func TestConfigSuite(t *testing.T) {
	test.RunSuite(t, new(ConfigSuite))
}