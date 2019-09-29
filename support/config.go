package support

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

// ConfigT offers a declarative way to map the environment variables.
type ConfigT struct {
	AppyEnv string `env:"APPY_ENV" envDefault:"development"`
	GoEnv   string `env:"GO_ENV" envDefault:"development"`

	// Server related configuration.
	HTTPDebugEnabled        bool          `env:"HTTP_DEBUG_ENABLED" envDefault:"false"`
	HTTPLogFilterParameters []string      `env:"HTTP_LOG_FILTER_PARAMETERS" envDefault:"password"`
	HTTPHealthCheckURL      string        `env:"HTTP_HEALTH_CHECK_URL" envDefault:"/health_check"`
	HTTPHost                string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	HTTPPort                string        `env:"HTTP_PORT" envDefault:"3000"`
	HTTPGracefulTimeout     time.Duration `env:"HTTP_GRACEFUL_TIMEOUT" envDefault:"30s"`
	HTTPIdleTimeout         time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"75s"`
	HTTPMaxHeaderBytes      int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"0"`
	HTTPReadTimeout         time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"60s"`
	HTTPReadHeaderTimeout   time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"60s"`
	HTTPWriteTimeout        time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"60s"`
	HTTPSSLCertPath         string        `env:"HTTP_SSL_CERT_PATH" envDefault:"./tmp/ssl"`
	HTTPSSLEnabled          bool          `env:"HTTP_SSL_ENABLED" envDefault:"false"`
	HTTPSSLPort             string        `env:"HTTP_SSL_PORT" envDefault:"3443"`

	// Session related configuration using cookie.
	HTTPSessionCookieDomain   string `env:"HTTP_SESSION_COOKIE_DOMAIN" envDefault:"0.0.0.0"`
	HTTPSessionCookieHTTPOnly bool   `env:"HTTP_SESSION_COOKIE_HTTP_ONLY" envDefault:"true"`
	HTTPSessionCookieMaxAge   int    `env:"HTTP_SESSION_COOKIE_MAX_AGE" envDefault:"0"`
	HTTPSessionCookiePath     string `env:"HTTP_SESSION_COOKIE_PATH" envDefault:"/"`
	HTTPSessionCookieSecure   bool   `env:"HTTP_SESSION_COOKIE_SECURE" envDefault:"true"`

	// Session related configuration using redis pool.
	HTTPSessionRedisAddr            string        `env:"HTTP_SESSION_REDIS_ADDR" envDefault:"0.0.0.0:6379"`
	HTTPSessionRedisAuth            string        `env:"HTTP_SESSION_REDIS_AUTH" envDefault:""`
	HTTPSessionRedisDb              string        `env:"HTTP_SESSION_REDIS_AUTH" envDefault:"0"`
	HTTPSessionRedisMaxActive       int           `env:"HTTP_SESSION_REDIS_MAX_ACTIVE" envDefault:"0"`
	HTTPSessionRedisMaxIdle         int           `env:"HTTP_SESSION_REDIS_MAX_IDLE" envDefault:"32"`
	HTTPSessionRedisIdleTimeout     time.Duration `env:"HTTP_SESSION_REDIS_IDLE_TIMEOUT" envDefault:"30s"`
	HTTPSessionRedisMaxConnLifetime time.Duration `env:"HTTP_SESSION_REDIS_MAX_CONN_LIFETIME" envDefault:"0s"`
	HTTPSessionRedisWait            bool          `env:"HTTP_SESSION_REDIS_WAIT" envDefault:"true"`

	// Session related configuration.
	HTTPSessionName     string   `env:"HTTP_SESSION_NAME" envDefault:"_session"`
	HTTPSessionProvider string   `env:"HTTP_SESSION_PROVIDER" envDefault:"cookie"`
	HTTPSessionSecrets  [][]byte `env:"HTTP_SESSION_SECRETS" envDefault:""`

	// Security related configuration.
	HTTPAllowedHosts            []string          `env:"HTTP_ALLOWED_HOSTS" envDefault:""`
	HTTPCSRFCookieDomain        string            `env:"HTTP_CSRF_COOKIE_DOMAIN" envDefault:"0.0.0.0"`
	HTTPCSRFCookieHTTPOnly      bool              `env:"HTTP_CSRF_COOKIE_HTTP_ONLY" envDefault:"true"`
	HTTPCSRFCookieMaxAge        int               `env:"HTTP_CSRF_COOKIE_MAX_AGE" envDefault:"43200"`
	HTTPCSRFCookieName          string            `env:"HTTP_CSRF_COOKIE_NAME" envDefault:"_csrf_token"`
	HTTPCSRFCookiePath          string            `env:"HTTP_CSRF_COOKIE_PATH" envDefault:"/"`
	HTTPCSRFCookieSecure        bool              `env:"HTTP_CSRF_COOKIE_SECURE" envDefault:"true"`
	HTTPCSRFFieldName           string            `env:"HTTP_CSRF_FIELD_NAME" envDefault:"authenticity_token"`
	HTTPCSRFRequestHeader       string            `env:"HTTP_CSRF_REQUEST_HEADER" envDefault:"X-CSRF-Token"`
	HTTPCSRFSecret              []byte            `env:"HTTP_CSRF_SECRET" envDefault:""`
	HTTPSSLRedirect             bool              `env:"HTTP_SSL_REDIRECT" envDefault:"false"`
	HTTPSSLTemporaryRedirect    bool              `env:"HTTP_SSL_TEMPORARY_REDIRECT" envDefault:"false"`
	HTTPSSLHost                 string            `env:"HTTP_SSL_HOST" envDefault:""`
	HTTPSTSSeconds              int64             `env:"HTTP_STS_SECONDS" envDefault:"0"`
	HTTPSTSIncludeSubdomains    bool              `env:"HTTP_STS_INCLUDE_SUBDOMAINS" envDefault:"false"`
	HTTPFrameDeny               bool              `env:"HTTP_FRAME_DENY" envDefault:"false"`
	HTTPCustomFrameOptionsValue string            `env:"HTTP_CUSTOM_FRAME_OPTIONS_VALUE" envDefault:""`
	HTTPContentTypeNosniff      bool              `env:"HTTP_CONTENT_TYPE_NOSNIFF" envDefault:"false"`
	HTTPBrowserXSSFilter        bool              `env:"HTTP_BROWSER_XSS_FILTER" envDefault:"false"`
	HTTPContentSecurityPolicy   string            `env:"HTTP_CONTENT_SECURITY_POLICY" envDefault:""`
	HTTPReferrerPolicy          string            `env:"HTTP_REFERRER_POLICY" envDefault:""`
	HTTPIENoOpen                bool              `env:"HTTP_IE_NO_OPEN" envDefault:"false"`
	HTTPSSLProxyHeaders         map[string]string `env:"HTTP_SSL_PROXY_HEADERS" envDefault:""`
}

var (
	// Build specifies if the binary is debug or release build.
	Build = "debug"

	// DotenvPath is the path to the current loaded .env file.
	DotenvPath = "None"
)

// NewConfig constructs a config.
func NewConfig() (*ConfigT, error) {
	dotenvPath := dotenvPath()
	if err := godotenv.Load(dotenvPath); err == nil {
		DotenvPath = dotenvPath
	}

	cfg := &ConfigT{}
	if err := ParseEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ParseEnv parses the environment variables into the config.
func ParseEnv(cfg interface{}) error {
	err := env.ParseWithFuncs(cfg, map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(map[string]string{}): func(v string) (interface{}, error) {
			newMaps := map[string]string{}
			maps := strings.Split(v, ",")
			for _, m := range maps {
				splits := strings.Split(m, ":")
				if len(splits) != 2 {
					continue
				}

				newMaps[splits[0]] = splits[1]
			}

			return newMaps, nil
		},
		reflect.TypeOf([]byte{}): func(v string) (interface{}, error) {
			return []byte(v), nil
		},
		reflect.TypeOf([][]byte{}): func(v string) (interface{}, error) {
			newBytes := [][]byte{}
			bytes := strings.Split(v, ",")
			for _, b := range bytes {
				newBytes = append(newBytes, []byte(b))
			}

			return newBytes, nil
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func dotenvPath() string {
	fn := ".env.development"
	if os.Getenv("APPY_ENV") != "" {
		fn = fmt.Sprintf(".env.%s", os.Getenv("APPY_ENV"))
	}

	return fn
}
