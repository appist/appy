package core

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/appist/appy/support"
	"github.com/joho/godotenv"
)

// AppConfig keeps the parsed environment variables.
type AppConfig struct {
	AppyEnv string `env:"APPY_ENV" envDefault:"development"`

	// Server related configuration.
	HTTPDebugEnabled        bool          `env:"HTTP_DEBUG_ENABLED" envDefault:"false"`
	HTTPLogFilterParameters []string      `env:"HTTP_LOG_FILTER_PARAMETERS" envDefault:"password"`
	HTTPHealthCheckURL      string        `env:"HTTP_HEALTH_CHECK_URL" envDefault:"/health_check"`
	HTTPHost                string        `env:"HTTP_HOST" envDefault:"localhost"`
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
	HTTPSessionCookieDomain   string `env:"HTTP_SESSION_COOKIE_DOMAIN" envDefault:"localhost"`
	HTTPSessionCookieHTTPOnly bool   `env:"HTTP_SESSION_COOKIE_HTTP_ONLY" envDefault:"true"`
	HTTPSessionCookieMaxAge   int    `env:"HTTP_SESSION_COOKIE_MAX_AGE" envDefault:"0"`
	HTTPSessionCookiePath     string `env:"HTTP_SESSION_COOKIE_PATH" envDefault:"/"`
	HTTPSessionCookieSecure   bool   `env:"HTTP_SESSION_COOKIE_SECURE" envDefault:"true"`

	// Session related configuration using redis pool.
	HTTPSessionRedisAddr            string        `env:"HTTP_SESSION_REDIS_ADDR" envDefault:"localhost:6379"`
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
	HTTPSessionSecrets  [][]byte `env:"HTTP_SESSION_SECRETS,required" envDefault:""`

	// Security related configuration.
	HTTPAllowedHosts            []string          `env:"HTTP_ALLOWED_HOSTS" envDefault:""`
	HTTPCSRFCookieDomain        string            `env:"HTTP_CSRF_COOKIE_DOMAIN" envDefault:"localhost"`
	HTTPCSRFCookieHTTPOnly      bool              `env:"HTTP_CSRF_COOKIE_HTTP_ONLY" envDefault:"true"`
	HTTPCSRFCookieMaxAge        int               `env:"HTTP_CSRF_COOKIE_MAX_AGE" envDefault:"43200"`
	HTTPCSRFCookieName          string            `env:"HTTP_CSRF_COOKIE_NAME" envDefault:"_csrf_token"`
	HTTPCSRFCookiePath          string            `env:"HTTP_CSRF_COOKIE_PATH" envDefault:"/"`
	HTTPCSRFCookieSecure        bool              `env:"HTTP_CSRF_COOKIE_SECURE" envDefault:"true"`
	HTTPCSRFFieldName           string            `env:"HTTP_CSRF_FIELD_NAME" envDefault:"authenticity_token"`
	HTTPCSRFRequestHeader       string            `env:"HTTP_CSRF_REQUEST_HEADER" envDefault:"X-CSRF-Token"`
	HTTPCSRFSecret              []byte            `env:"HTTP_CSRF_SECRET,required" envDefault:""`
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

	// CSRPaths specifies the path that stores the client-side rendering assets.
	CSRPaths = map[string]string{
		"root": "web",
	}

	// SSRPaths specifies the paths that store the server-side rendering assets.
	SSRPaths = map[string]string{
		"root":   ".ssr",
		"config": "config",
		"locale": "app/locales",
		"view":   "app/views",
	}

	staticExtRegex = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
)

func getConfigInfo(assets http.FileSystem) (string, io.Reader, error) {
	var (
		err    error
		reader io.Reader
	)

	if os.Getenv("APPY_ENV") == "" {
		os.Setenv("APPY_ENV", "development")
	}

	path := SSRPaths["config"] + "/.env." + os.Getenv("APPY_ENV")
	if Build == "debug" {
		reader, err = os.Open(path)
	} else {
		path = SSRPaths["root"] + "/" + path

		if assets != nil {
			reader, err = assets.Open(path)
		}
	}

	if err != nil {
		path = "none"
	}

	return path, reader, err
}

func newConfig(assets http.FileSystem) (AppConfig, error) {
	_, reader, err := getConfigInfo(assets)
	if err == nil {
		envMap, _ := godotenv.Parse(reader)
		currentEnv := map[string]bool{}
		rawEnv := os.Environ()
		for _, rawEnvLine := range rawEnv {
			key := strings.Split(rawEnvLine, "=")[0]
			currentEnv[key] = true
		}

		// Add decrypt using APPY_MASTER_KEY
		for key, value := range envMap {
			if !currentEnv[key] {
				os.Setenv(key, value)
			}
		}
	}

	c := &AppConfig{}
	err = support.ParseEnv(c)

	return *c, err
}

// MasterKey retrieves the encryption/decryption key by checking the below in order:
//
// 1. the key in `config/<APPY_ENV>.key`
// 2. `APPY_MASTER_KEY` environment variable
func MasterKey() ([]byte, error) {
	appyEnv := "development"
	if os.Getenv("APPY_ENV") != "" {
		appyEnv = os.Getenv("APPY_ENV")
	}

	key, err := ioutil.ReadFile(SSRPaths["config"] + "/" + appyEnv + ".key")
	if err != nil {
		return nil, err
	}

	if os.Getenv("APPY_MASTER_KEY") != "" {
		key = []byte(os.Getenv("APPY_MASTER_KEY"))
	}

	key = []byte(strings.Trim(string(key), "\n"))
	key = []byte(strings.Trim(string(key), " "))

	if len(key) == 0 {
		return nil, errors.New("the master key cannot be blank, please either pass in \"APPY_MASTER_KEY\" environment variable or store it in \"config/<APPY_ENV>.key\"")
	}

	return key, nil
}
