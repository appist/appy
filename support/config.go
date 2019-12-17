package support

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type (
	// Config defines the application settings.
	Config struct {
		AppyEnv string `env:"APPY_ENV" envDefault:"development"`

		// GraphQL related configuration.
		GQLPlaygroundEnabled          bool          `env:"GQL_PLAYGROUND_ENABLED" envDefault:"false"`
		GQLPlaygroundPath             string        `env:"GQL_PLAYGROUND_PATH" envDefault:"/docs/graphql"`
		GQLCacheSize                  int           `env:"GQL_CACHE_SIZE" envDefault:"1000"`
		GQLComplexityLimit            int           `env:"GQL_COMPLEXITY_LIMIT" envDefault:"200"`
		GQLUploadMaxMemory            int64         `env:"GQL_UPLOAD_MAX_MEMORY" envDefault:"100000000"`
		GQLUploadMaxSize              int64         `env:"GQL_UPLOAD_MAX_SIZE" envDefault:"100000000"`
		GQLWebsocketKeepAliveDuration time.Duration `env:"GQL_WEBSOCKET_KEEP_ALIVE_DURATION" envDefault:"30s"`

		// Server related configuration.
		HTTPDebugEnabled        bool          `env:"HTTP_DEBUG_ENABLED" envDefault:"false"`
		HTTPGzipCompressLevel   int           `env:"HTTP_GZIP_COMPRESS_LEVEL" envDefault:"-1"`
		HTTPGzipExcludedExts    []string      `env:"HTTP_GZIP_EXCLUDED_EXTS" envDefault:""`
		HTTPGzipExcludedPaths   []string      `env:"HTTP_GZIP_EXCLUDED_PATHS" envDefault:""`
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

		// Session related configuration using redis pool.
		HTTPSessionRedisAddr            string        `env:"HTTP_SESSION_REDIS_ADDR" envDefault:"localhost:6379"`
		HTTPSessionRedisAuth            string        `env:"HTTP_SESSION_REDIS_AUTH" envDefault:""`
		HTTPSessionRedisDb              string        `env:"HTTP_SESSION_REDIS_DB" envDefault:"0"`
		HTTPSessionRedisMaxActive       int           `env:"HTTP_SESSION_REDIS_MAX_ACTIVE" envDefault:"64"`
		HTTPSessionRedisMaxIdle         int           `env:"HTTP_SESSION_REDIS_MAX_IDLE" envDefault:"32"`
		HTTPSessionRedisIdleTimeout     time.Duration `env:"HTTP_SESSION_REDIS_IDLE_TIMEOUT" envDefault:"30s"`
		HTTPSessionRedisMaxConnLifetime time.Duration `env:"HTTP_SESSION_REDIS_MAX_CONN_LIFETIME" envDefault:"30s"`
		HTTPSessionRedisWait            bool          `env:"HTTP_SESSION_REDIS_WAIT" envDefault:"true"`

		// Session related configuration.
		HTTPSessionName       string   `env:"HTTP_SESSION_NAME" envDefault:"_session"`
		HTTPSessionProvider   string   `env:"HTTP_SESSION_PROVIDER" envDefault:"cookie"`
		HTTPSessionExpiration int      `env:"HTTP_SESSION_EXPIRATION" envDefault:"1209600"`
		HTTPSessionDomain     string   `env:"HTTP_SESSION_DOMAIN" envDefault:"localhost"`
		HTTPSessionHTTPOnly   bool     `env:"HTTP_SESSION_HTTP_ONLY" envDefault:"true"`
		HTTPSessionPath       string   `env:"HTTP_SESSION_PATH" envDefault:"/"`
		HTTPSessionSecure     bool     `env:"HTTP_SESSION_SECURE" envDefault:"false"`
		HTTPSessionSecrets    [][]byte `env:"HTTP_SESSION_SECRETS,required" envDefault:""`

		// Security related configuration.
		HTTPAllowedHosts            []string          `env:"HTTP_ALLOWED_HOSTS" envDefault:""`
		HTTPCSRFCookieDomain        string            `env:"HTTP_CSRF_COOKIE_DOMAIN" envDefault:"localhost"`
		HTTPCSRFCookieHTTPOnly      bool              `env:"HTTP_CSRF_COOKIE_HTTP_ONLY" envDefault:"true"`
		HTTPCSRFCookieMaxAge        int               `env:"HTTP_CSRF_COOKIE_MAX_AGE" envDefault:"0"`
		HTTPCSRFCookieName          string            `env:"HTTP_CSRF_COOKIE_NAME" envDefault:"_csrf_token"`
		HTTPCSRFCookiePath          string            `env:"HTTP_CSRF_COOKIE_PATH" envDefault:"/"`
		HTTPCSRFCookieSecure        bool              `env:"HTTP_CSRF_COOKIE_SECURE" envDefault:"false"`
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
		HTTPSSLProxyHeaders         map[string]string `env:"HTTP_SSL_PROXY_HEADERS" envDefault:"X-Forwarded-Proto:https"`

		// Mailer related configuration.
		MailerSMTPAddr              string `env:"MAILER_SMTP_ADDR" envDefault:""`
		MailerSMTPPlainAuthIdentity string `env:"MAILER_SMTP_PLAIN_AUTH_IDENTITY" envDefault:""`
		MailerSMTPPlainAuthUsername string `env:"MAILER_SMTP_PLAIN_AUTH_USERNAME" envDefault:""`
		MailerSMTPPlainAuthPassword string `env:"MAILER_SMTP_PLAIN_AUTH_PASSWORD" envDefault:""`
		MailerSMTPPlainAuthHost     string `env:"MAILER_SMTP_PLAIN_AUTH_HOST" envDefault:""`
		MailerPreviewBaseURL        string `env:"MAILER_PREVIEW_BASE_URL" envDefault:"/appy/mailers"`

		path      string
		errors    []error
		masterKey []byte
	}
)

// NewConfig initializes Config instance.
func NewConfig(assetsMngr *AssetsMngr, logger *Logger) *Config {
	var (
		errs []error
	)

	masterKey, err := parseMasterKey(assetsMngr)
	if err != nil {
		errs = append(errs, err)
	}

	config := &Config{}
	if masterKey != nil {
		config.path = assetsMngr.Layout()["config"] + "/.env." + os.Getenv("APPY_ENV")
		config.masterKey = masterKey
		decryptErrs := config.decryptConfig(assetsMngr, masterKey)
		if len(decryptErrs) > 0 {
			errs = append(errs, decryptErrs...)
		}

		err = ParseEnv(config)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		config.errors = errs
	}

	return config
}

// MasterKey returns the master key for the current environment.
func (c Config) MasterKey() []byte {
	return c.masterKey
}

// Errors returns all the config retrieving/parsing errors.
func (c Config) Errors() []error {
	return c.errors
}

// Path returns the config path.
func (c Config) Path() string {
	return c.path
}

func (c Config) decryptConfig(assetsMngr *AssetsMngr, masterKey []byte) []error {
	reader, err := assetsMngr.Open(c.path)
	if err != nil {
		return []error{err}
	}

	envMap, err := godotenv.Parse(reader)
	if err != nil {
		return []error{err}
	}

	// Parse the environment variables that aren't defined in the config .env file which can be used to override the ones
	// defined in config .env file.
	origEnvMap := map[string]bool{}
	origEnv := os.Environ()
	for _, rawEnvLine := range origEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		origEnvMap[key] = true
	}

	var errs []error
	if len(masterKey) != 0 {
		for key, value := range envMap {
			if origEnvMap[key] {
				continue
			}

			decodeStr, _ := hex.DecodeString(value)
			plaintext, err := AESDecrypt(decodeStr, masterKey)
			if len(plaintext) < 1 || err != nil {
				errs = append(errs, fmt.Errorf("unable to decrypt '%s' value in '%s'", key, c.path))
			}

			os.Setenv(key, string(plaintext))
		}
	}

	return errs
}

// IsProtectedEnv is used to protect the app from being destroyed by a command accidentally.
func IsProtectedEnv(config *Config) bool {
	return config.AppyEnv == "production"
}

func parseMasterKey(assetsMngr *AssetsMngr) ([]byte, error) {
	var (
		err error
		key []byte
	)

	env := "development"
	if os.Getenv("APPY_ENV") != "" {
		env = os.Getenv("APPY_ENV")
	}

	if os.Getenv("APPY_MASTER_KEY") != "" {
		key = []byte(os.Getenv("APPY_MASTER_KEY"))
	}

	if len(key) == 0 && IsDebugBuild() {
		key, err = ioutil.ReadFile(assetsMngr.Layout()["config"] + "/" + env + ".key")
		if err != nil {
			return nil, ErrReadMasterKeyFile
		}
	}

	key = []byte(strings.Trim(string(key), "\n"))
	key = []byte(strings.Trim(string(key), " "))

	if len(key) == 0 {
		return nil, ErrNoMasterKey
	}

	return key, nil
}
