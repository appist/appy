package appy

import (
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
)

type ConfigSuite struct {
	TestSuite
	oldSSRPaths map[string]string
}

func (s *ConfigSuite) SetupTest() {
	s.oldSSRPaths = _ssrPaths
	_ssrPaths = map[string]string{
		"root":   "testdata/.ssr",
		"config": "testdata/pkg/config",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
	}
}

func (s *ConfigSuite) TearDownTest() {
	_ssrPaths = s.oldSSRPaths
}

func (s *ConfigSuite) TestNewConfigDefaultValue() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")

	tt := map[string]interface{}{
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
		"HTTPSessionCookieSecure":         false,
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
		"HTTPSSLProxyHeaders":             map[string]string{},
	}

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
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

	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithoutSettingRequiredConfig() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	s.NotNil(config.Errors())
	s.EqualError(config.Errors()[0], `required environment variable "HTTP_SESSION_SECRETS" is not set. required environment variable "HTTP_CSRF_SECRET" is not set`)

	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithSettingRequiredConfig() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	s.Equal([]byte("481e5d98a31585148b8b1dfb6a3c0465"), config.MasterKey())
	s.Nil(config.Errors())

	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ConfigSuite) TestNewConfigWithUnparsableEnvVariable() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_DEBUG_ENABLED", "nil")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	s.Contains(config.Errors()[0].Error(), `strconv.ParseBool: parsing "nil": invalid syntax.`)

	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_DEBUG_ENABLED")
}

func (s *ConfigSuite) TestNewConfigWithUndecryptableConfig() {
	oldAppyEnv := os.Getenv("APPY_ENV")
	os.Setenv("APPY_ENV", "undecryptable")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	s.Contains(config.Errors()[0].Error(), "unable to decrypt 'HTTP_PORT' value in 'testdata/pkg/config/.env.undecryptable'")

	os.Setenv("APPY_ENV", oldAppyEnv)
	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithInvalidAssetsPath() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")

	build := ReleaseBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, http.Dir("testdata"))
	s.Contains(config.Errors()[0].Error(), "open testdata/testdata/.ssr/testdata/pkg/config/.env.development: no such file or directory")

	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithMissingConfigInAssets() {
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")

	build := ReleaseBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	s.EqualError(config.Errors()[0], ErrNoConfigInAssets.Error())

	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithUnparsableConfig() {
	oldAppyEnv := os.Getenv("APPY_ENV")
	os.Setenv("APPY_ENV", "unparsable")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	s.Contains(config.Errors()[0].Error(), "Can't separate key from value")

	os.Setenv("APPY_ENV", oldAppyEnv)
	os.Unsetenv("APPY_MASTER_KEY")
}

func (s *ConfigSuite) TestNewConfigWithInvalidDatabaseConfig() {
	oldAppyEnv := os.Getenv("APPY_ENV")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("DB_ADDR_PRIMARY", "dummy")
	os.Setenv("DB_APP_NAME_PRIMARY", "dummy")
	os.Setenv("DB_DIAL_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_IDLE_CHECK_FREQUENCY_PRIMARY", "dummy")
	os.Setenv("DB_IDLE_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_MAX_CONN_AGE_PRIMARY", "dummy")
	os.Setenv("DB_MAX_RETRIES_PRIMARY", "dummy")
	os.Setenv("DB_MIN_IDLE_CONNS_PRIMARY", "dummy")
	os.Setenv("DB_PASSWORD_PRIMARY", "dummy")
	os.Setenv("DB_POOL_SIZE_PRIMARY", "dummy")
	os.Setenv("DB_POOL_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_READ_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("DB_REPLICA_PRIMARY", "dummy")
	os.Setenv("DB_RETRY_STATEMENT_PRIMARY", "dummy")
	os.Setenv("DB_SCHEMA_SEARCH_PATH_PRIMARY", "dummy")
	os.Setenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY", "true")
	os.Setenv("DB_USER_PRIMARY", "dummy")
	os.Setenv("DB_WRITE_TIMEOUT_PRIMARY", "dummy")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	dbManager := NewDbManager(logger)
	s.Nil(config.Errors())
	s.NotNil(dbManager.Errors())

	os.Setenv("APPY_ENV", oldAppyEnv)
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("DB_ADDR_PRIMARY")
	os.Unsetenv("DB_APP_NAME_PRIMARY")
	os.Unsetenv("DB_DIAL_TIMEOUT_PRIMARY")
	os.Unsetenv("DB_IDLE_CHECK_FREQUENCY_PRIMARY")
	os.Unsetenv("DB_IDLE_TIMEOUT_PRIMARY")
	os.Unsetenv("DB_MAX_CONN_AGE_PRIMARY")
	os.Unsetenv("DB_MAX_RETRIES_PRIMARY")
	os.Unsetenv("DB_MIN_IDLE_CONNS_PRIMARY")
	os.Unsetenv("DB_PASSWORD_PRIMARY")
	os.Unsetenv("DB_POOL_SIZE_PRIMARY")
	os.Unsetenv("DB_POOL_TIMEOUT_PRIMARY")
	os.Unsetenv("DB_READ_TIMEOUT_PRIMARY")
	os.Unsetenv("DB_REPLICA_PRIMARY")
	os.Unsetenv("DB_RETRY_STATEMENT_PRIMARY")
	os.Unsetenv("DB_SCHEMA_SEARCH_PATH_PRIMARY")
	os.Unsetenv("DB_SCHEMA_MIGRATIONS_TABLE_PRIMARY")
	os.Unsetenv("DB_USER_PRIMARY")
	os.Unsetenv("DB_WRITE_TIMEOUT_PRIMARY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ConfigSuite) TestNewConfigWithValidDatabaseConfig() {
	oldAppyEnv := os.Getenv("APPY_ENV")
	os.Setenv("APPY_ENV", "valid_db")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	build := DebugBuild
	logger := NewLogger(build)
	config := NewConfig(build, logger, nil)
	dbManager := NewDbManager(logger)
	s.Nil(config.Errors())
	s.Nil(dbManager.Errors())
	s.Nil(dbManager.ConnectAll(true))

	os.Setenv("APPY_ENV", oldAppyEnv)
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("DB_ADDR_PRIMARY")
	os.Unsetenv("DB_NAME_PRIMARY")
	os.Unsetenv("DB_PASSWORD_PRIMARY")
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
	os.Setenv("APPY_MASTER_KEY", "")
	Build = ReleaseBuild
	_, err := parseMasterKey()
	s.EqualError(err, ErrNoMasterKey.Error())
	os.Unsetenv("APPY_MASTER_KEY")
	Build = DebugBuild
}

func TestConfigSuite(t *testing.T) {
	RunTestSuite(t, new(ConfigSuite))
}
