package support

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/appist/appy/test"
)

func TestNewConfig(t *testing.T) {
	assert := test.NewAssert(t)

	t.Run("sets default values with correct type", func(t *testing.T) {
		tests := map[string]interface{}{
			"AppyEnv":                         "development",
			"GoEnv":                           "development",
			"HTTPDebugEnabled":                false,
			"HTTPLogFilterParameters":         []string{"password"},
			"HTTPHealthCheckURL":              "/health_check",
			"HTTPHost":                        "0.0.0.0",
			"HTTPPort":                        "3000",
			"HTTPGracefulTimeout":             30 * time.Second,
			"HTTPIdleTimeout":                 75 * time.Second,
			"HTTPMaxHeaderBytes":              0,
			"HTTPReadTimeout":                 60 * time.Second,
			"HTTPReadHeaderTimeout":           60 * time.Second,
			"HTTPWriteTimeout":                60 * time.Second,
			"HTTPSSLEnabled":                  false,
			"HTTPSSLCertPath":                 "./tmp/ssl",
			"HTTPSessionCookieDomain":         "0.0.0.0",
			"HTTPSessionCookieHTTPOnly":       true,
			"HTTPSessionCookieMaxAge":         0,
			"HTTPSessionCookiePath":           "/",
			"HTTPSessionCookieSecure":         true,
			"HTTPSessionRedisAddr":            "0.0.0.0:6379",
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
			"HTTPCSRFCookieDomain":            "0.0.0.0",
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

		c, _ := NewConfig()
		cv := reflect.ValueOf(*c)
		for key, defaultVal := range tests {
			fv := cv.FieldByName(key)

			// An exception case to handle a different host in test for Github actions.
			if key == "HTTPSessionRedisAddr" && os.Getenv("HTTP_SESSION_REDIS_ADDR") != "" {
				assert.Equal(fv.Interface(), os.Getenv("HTTP_SESSION_REDIS_ADDR"))
				continue
			}

			switch fv.Kind() {
			case reflect.Map:
				switch fv.Type().String() {
				case "map[string]string":
					for key, val := range fv.Interface().(map[string]string) {
						assert.Equal(val, defaultVal.(map[string]string)[key])
					}
				}
			case reflect.Slice, reflect.Array:
				switch fv.Type().String() {
				case "[]string":
					assert.Equal(len(fv.Interface().([]string)), len(defaultVal.([]string)))

					for idx, val := range fv.Interface().([]string) {
						assert.Equal(val, defaultVal.([]string)[idx])
					}
				case "[]uint8":
					assert.Equal(len(fv.Interface().([]uint8)), len(defaultVal.([]uint8)))

					for idx, val := range fv.Interface().([]uint8) {
						assert.Equal(val, defaultVal.([]uint8)[idx])
					}
				case "[][]uint8":
					assert.Equal(len(fv.Interface().([][]uint8)), len(defaultVal.([][]uint8)))

					for idx, val := range fv.Interface().([][]uint8) {
						assert.Equal(val, defaultVal.([][]uint8)[idx])
					}
				default:
					assert.Equal(fv.Interface(), defaultVal)
				}
			default:
				assert.Equal(fv.Interface(), defaultVal)
			}
		}
	})

	t.Run("returns error if the environment variable cannot be parsed", func(t *testing.T) {
		os.Setenv("HTTP_DEBUG_ENABLED", "nil")
		_, err := NewConfig()
		os.Unsetenv("HTTP_DEBUG_ENABLED")
		assert.NotNil(err)
	})
}

func TestParseEnv(t *testing.T) {
	assert := test.NewAssert(t)

	t.Run("parses the environment variables correctly", func(t *testing.T) {
		type config struct {
			Admins  map[string]string `env:"TEST_ADMINS" envDefault:"user1:pass1,user2:pass2"`
			Hosts   []string          `env:"TEST_HOSTS" envDefault:"0.0.0.0,1.1.1.1"`
			Secret  []byte            `env:"TEST_SECRET" envDefault:"hello"`
			Secrets [][]byte          `env:"TEST_SECRETS" envDefault:"hello,world"`
		}

		cfg := &config{}
		ParseEnv(cfg)
		assert.Equal(map[string]string{"user1": "pass1", "user2": "pass2"}, cfg.Admins)
		assert.Equal([]string{"0.0.0.0", "1.1.1.1"}, cfg.Hosts)
		assert.Equal([]byte("hello"), cfg.Secret)
		assert.Equal([][]byte{[]byte("hello"), []byte("world")}, cfg.Secrets)
	})

	t.Run("returns error if the type isn't supported", func(t *testing.T) {
		type config struct {
			Users map[string]int `env:"TEST_USERS" envDefault:"user1:1,user2:2"`
		}

		cfg := &config{}
		err := ParseEnv(cfg)
		assert.NotNil(err)
	})

	t.Run("returns empty map if the environment variable for map[string]string", func(t *testing.T) {
		type config struct {
			Users map[string]string `env:"TEST_USERS" envDefault:"user1"`
		}

		cfg := &config{}
		ParseEnv(cfg)
		assert.Equal(map[string]string{}, cfg.Users)
	})
}

func TestDotenv(t *testing.T) {
	assert := test.NewAssert(t)
	assert.Equal(".env.development", dotenvPath())

	os.Setenv("APPY_ENV", "production")
	dotenvPath := dotenvPath()
	os.Unsetenv("APPY_ENV")
	assert.Equal(".env.production", dotenvPath)
}
