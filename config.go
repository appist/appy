package appy

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type (
	// Config defines the application settings.
	Config struct {
		// AppyEnv indicates the environment that the codebase is running on and it determines which config to use. By
		// default, it is is "development" and its corresponding config is "configs/.env.development".
		//
		// Note: APPY_ENV=test is used for unit tests.
		AppyEnv string `env:"APPY_ENV" envDefault:"development"`

		// AssetHost indicates the asset host to use with "assetPath()" for the server-side rendering which is very useful
		// it comes to hosting the static assets on CDN. By default, it is "" which uses the current server host.
		AssetHost string `env:"ASSET_HOST" envDefault:""`

		// GQLPlaygroundEnabled indicates if the GraphQL playground is enabled. By default, it is false.
		GQLPlaygroundEnabled bool `env:"GQL_PLAYGROUND_ENABLED" envDefault:"false"`

		// GQLPlaygroundPath indicates the GraphQL playground path to host at. By default, it is "/docs/graphql".
		GQLPlaygroundPath string `env:"GQL_PLAYGROUND_PATH" envDefault:"/docs/graphql"`

		// GQLAPQCacheSize indicates how many APQ to persist in the memory at one time. By default, it is 100.
		//
		// For more details about APQ, please refer to https://gqlgen.com/reference/apq.
		GQLAPQCacheSize int `env:"GQL_APQ_CACHE_SIZE" envDefault:"100"`

		// GQLQueryCacheSize indicates how many queries to cache in the memory. By default, it is 1000.
		GQLQueryCacheSize int `env:"GQL_QUERY_CACHE_SIZE" envDefault:"1000"`

		// GQLComplexityLimit indicates the query complexity which can be used to mitigate DDoS attacks risk. By default,
		// it is 1000.
		GQLComplexityLimit int `env:"GQL_COMPLEXITY_LIMIT" envDefault:"1000"`

		// GQLMultipartMaxMemory indicates the maximum number of bytes used to parse a request body as multipart/form-data
		// in memory, with the remainder stored on disk in temporary files. By default, it is 0 (no limit).
		GQLMultipartMaxMemory int64 `env:"GQL_MULTIPART_MAX_MEMORY" envDefault:"0"`

		// GQLMultipartMaxUploadSize indicates the maximum number of bytes used to parse a request body as
		// multipart/form-data. By default, it is 0 (no limit).
		GQLMultipartMaxUploadSize int64 `env:"GQL_MULTIPART_MAX_UPLOAD_SIZE" envDefault:"0"`

		// GQLWebsocketKeepAliveDuration indicates how long the websocket connection should be kept alive for sending
		// subsequent messages without re-establishing the connection which is an overhead. By default, it is 10s.
		GQLWebsocketKeepAliveDuration time.Duration `env:"GQL_WEBSOCKET_KEEP_ALIVE_DURATION" envDefault:"10s"`

		// HTTPGzipCompressLevel indicates the compression level used to compress the HTTP response. By default, it is -1.
		//
		// Available compression level:
		// 	 - Default Compression -> -1
		//   - No Compression -> 0
		//   - Fastest Compression -> 1
		//   - Best Compression -> 9
		HTTPGzipCompressLevel int `env:"HTTP_GZIP_COMPRESS_LEVEL" envDefault:"-1"`

		// HTTPGzipExcludedExts indicates which file extensions not to compress. By default, it is "".
		HTTPGzipExcludedExts []string `env:"HTTP_GZIP_EXCLUDED_EXTS" envDefault:""`

		// HTTPGzipExcludedPaths indicates which paths not to compress. By default, it is "".
		HTTPGzipExcludedPaths []string `env:"HTTP_GZIP_EXCLUDED_PATHS" envDefault:""`

		// HTTPLogFilterParameters indicates which query parameters in the URL to filter so that the sensitive information
		// like password are masked in the HTTP request log. By default, it is "password".
		HTTPLogFilterParameters []string `env:"HTTP_LOG_FILTER_PARAMETERS" envDefault:"password"`

		// HTTPHealthCheckURL indicates the path to check if the HTTP server is healthy. This endpoint is a middleware
		// that is designed to avoid redundant computing resource usage. By default, it is "/health_check".
		//
		// In general, if your server is running behind a load balancer, this endpoint will be served to inform the load
		// balancer that the server is healthy and ready to receive HTTP requests.
		HTTPHealthCheckURL string `env:"HTTP_HEALTH_CHECK_URL" envDefault:"/health_check"`

		// HTTPHost indicates which host the HTTP server should be hosted at. By default, it is "localhost". If you would
		// like to connect to the HTTP server from within your LAN network, use "0.0.0.0" instead.
		HTTPHost string `env:"HTTP_HOST" envDefault:"localhost"`

		// HTTPPort indicates which port the HTTP server should be hosted at. By default, it is "3000".
		HTTPPort string `env:"HTTP_PORT" envDefault:"3000"`

		// HTTPGracefulTimeout indicates how long to wait for the HTTP server to shut down so that any active connection
		// is not interrupted by SIGTERM/SIGINT. By default, it is "30s".
		HTTPGracefulTimeout time.Duration `env:"HTTP_GRACEFUL_TIMEOUT" envDefault:"30s"`

		// HTTPIdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled. If
		// HTTPIdleTimeout is zero, the value of HTTPReadTimeout is used. If both are zero, there is no timeout. By
		// default, it is "75s".
		HTTPIdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"75s"`

		// HTTPMaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header's keys
		// and values, including the request line. It does not limit the size of the request body. If zero,
		// http.DefaultMaxHeaderBytes (1 << 20 which is 1 MB) is used.
		HTTPMaxHeaderBytes int `env:"HTTP_MAX_HEADER_BYTES" envDefault:"0"`

		// HTTPReadTimeout is the maximum duration for reading the entire request, including the body. Because
		// HTTPReadTimeout does not let Handlers make per-request decisions on each request body's acceptable deadline or
		// upload rate, most users will prefer to use HTTPReadHeaderTimeout. It is valid to use them both. By default, it
		// is "60s".
		HTTPReadTimeout time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"60s"`

		// HTTPReadHeaderTimeout is the amount of time allowed to read request headers. The connection's read deadline is
		// reset after reading the headers and the Handler can decide what is considered too slow for the body. If
		// HTTPReadHeaderTimeout is zero, the value of HTTPReadTimeout is used. If both are zero, there is no timeout. By
		// default, it is "60s".
		HTTPReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"60s"`

		// HTTPWriteTimeout is the maximum duration before timing out writes of the response. It is reset whenever a new
		// request's header is read. Like HTTPReadTimeout, it does not let Handlers make decisions on a per-request basis.
		// By default, it is "60s".
		HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"60s"`

		// HTTPSSLCertPath indicates which path to store the locally trusted SSL certificates which are created using
		// "go run . ssl:setup" command. By default, it is "./tmp/ssl".
		HTTPSSLCertPath string `env:"HTTP_SSL_CERT_PATH" envDefault:"./tmp/ssl"`

		// HTTPSSLEnabled indicates if the HTTPS server should be enabled. When enabled, please ensure the SSL certificates
		// are created in the "HTTPSSLCertPath". By default, it is false.
		HTTPSSLEnabled bool `env:"HTTP_SSL_ENABLED" envDefault:"false"`

		// HTTPSSLPort indicates which port the HTTPS server should be hosted at. By default, it is "3443".
		HTTPSSLPort string `env:"HTTP_SSL_PORT" envDefault:"3443"`

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
		HTTPSessionName       string        `env:"HTTP_SESSION_NAME" envDefault:"_session"`
		HTTPSessionProvider   string        `env:"HTTP_SESSION_PROVIDER" envDefault:"cookie"`
		HTTPSessionExpiration int           `env:"HTTP_SESSION_EXPIRATION" envDefault:"1209600"`
		HTTPSessionDomain     string        `env:"HTTP_SESSION_DOMAIN" envDefault:"localhost"`
		HTTPSessionHTTPOnly   bool          `env:"HTTP_SESSION_HTTP_ONLY" envDefault:"true"`
		HTTPSessionPath       string        `env:"HTTP_SESSION_PATH" envDefault:"/"`
		HTTPSessionSameSite   http.SameSite `env:"HTTP_SESSION_SAME_SITE" envDefault:"1"`
		HTTPSessionSecure     bool          `env:"HTTP_SESSION_SECURE" envDefault:"false"`
		HTTPSessionSecrets    [][]byte      `env:"HTTP_SESSION_SECRETS,required" envDefault:""`

		// Security related configuration.
		HTTPAllowedHosts            []string          `env:"HTTP_ALLOWED_HOSTS" envDefault:""`
		HTTPCSRFCookieDomain        string            `env:"HTTP_CSRF_COOKIE_DOMAIN" envDefault:"localhost"`
		HTTPCSRFCookieHTTPOnly      bool              `env:"HTTP_CSRF_COOKIE_HTTP_ONLY" envDefault:"true"`
		HTTPCSRFCookieMaxAge        int               `env:"HTTP_CSRF_COOKIE_MAX_AGE" envDefault:"0"`
		HTTPCSRFCookieName          string            `env:"HTTP_CSRF_COOKIE_NAME" envDefault:"_csrf_token"`
		HTTPCSRFCookiePath          string            `env:"HTTP_CSRF_COOKIE_PATH" envDefault:"/"`
		HTTPCSRFCookieSameSite      http.SameSite     `env:"HTTP_CSRF_COOKIE_SAME_SITE" envDefault:"1"`
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

		// I18nDefaultLocale indicates the default locale to use for translations in handlers/mailers/views when the desire
		// locale is not found. By default, it is "en".
		//
		// Note: If the locale is "en", the translation file would be "pkg/locales/en.yml".
		I18nDefaultLocale string `env:"I18N_DEFAULT_LOCALE" envDefault:"en"`

		// Mailer related configuration.
		MailerSMTPAddr              string `env:"MAILER_SMTP_ADDR" envDefault:""`
		MailerSMTPPlainAuthIdentity string `env:"MAILER_SMTP_PLAIN_AUTH_IDENTITY" envDefault:""`
		MailerSMTPPlainAuthUsername string `env:"MAILER_SMTP_PLAIN_AUTH_USERNAME" envDefault:""`
		MailerSMTPPlainAuthPassword string `env:"MAILER_SMTP_PLAIN_AUTH_PASSWORD" envDefault:""`
		MailerSMTPPlainAuthHost     string `env:"MAILER_SMTP_PLAIN_AUTH_HOST" envDefault:""`
		MailerPreviewBaseURL        string `env:"MAILER_PREVIEW_BASE_URL" envDefault:"/appy/mailers"`

		// Worker related configuration.
		WorkerRedisSentinelAddrs      []string       `env:"WORKER_REDIS_SENTINEL_ADDRS" envDefault:""`
		WorkerRedisSentinelDB         int            `env:"WORKER_REDIS_SENTINEL_DB" envDefault:"0"`
		WorkerRedisSentinelMasterName string         `env:"WORKER_REDIS_SENTINEL_MASTER_NAME" envDefault:""`
		WorkerRedisSentinelPassword   string         `env:"WORKER_REDIS_SENTINEL_PASSWORD" envDefault:""`
		WorkerRedisSentinelPoolSize   int            `env:"WORKER_REDIS_SENTINEL_POOL_SIZE" envDefault:"25"`
		WorkerRedisAddr               string         `env:"WORKER_REDIS_ADDR" envDefault:""`
		WorkerRedisDB                 int            `env:"WORKER_REDIS_DB" envDefault:"0"`
		WorkerRedisPassword           string         `env:"WORKER_REDIS_PASSWORD" envDefault:""`
		WorkerRedisPoolSize           int            `env:"WORKER_REDIS_POOL_SIZE" envDefault:"25"`
		WorkerRedisURL                string         `env:"WORKER_REDIS_URL" envDefault:""`
		WorkerConcurrency             int            `env:"WORKER_CONCURRENCY" envDefault:"25"`
		WorkerQueues                  map[string]int `env:"WORKER_QUEUES" envDefault:"default:10"`
		WorkerStrictPriority          bool           `env:"WORKER_STRICT_PRIORITY" envDefault:"false"`

		path      string
		errors    []error
		masterKey []byte
	}
)

// NewConfig initializes Config instance.
func NewConfig(asset *Asset, logger *Logger, support Supporter) *Config {
	var errs []error

	masterKey, err := parseMasterKey(asset)
	if err != nil {
		errs = append(errs, err)
	}

	config := &Config{}
	if masterKey != nil {
		config.path = asset.Layout()["config"] + "/.env." + os.Getenv("APPY_ENV")
		config.masterKey = masterKey
		decryptErrs := config.decryptConfig(asset, masterKey, support)
		if len(decryptErrs) > 0 {
			errs = append(errs, decryptErrs...)
		}

		err = support.ParseEnv(config)
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
func (c *Config) MasterKey() []byte {
	return c.masterKey
}

// Errors returns all the config retrieving/parsing errors.
func (c *Config) Errors() []error {
	return c.errors
}

// Path returns the config path.
func (c *Config) Path() string {
	return c.path
}

// IsProtectedEnv is used to protect the app from being destroyed by a command accidentally.
func (c *Config) IsProtectedEnv() bool {
	return c.AppyEnv == "production"
}

func (c Config) decryptConfig(asset *Asset, masterKey []byte, support Supporter) []error {
	reader, err := asset.Open(c.path)
	if err != nil {
		return []error{err}
	}

	envMap, err := godotenv.Parse(reader)
	if err != nil {
		os.Clearenv()
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
			if origEnvMap[key] || value == "" {
				continue
			}

			decodeStr, _ := hex.DecodeString(value)
			plaintext, err := support.AESDecrypt(decodeStr, masterKey)
			if len(plaintext) < 1 || err != nil {
				errs = append(errs, fmt.Errorf("unable to decrypt '%s' value in '%s'", key, c.path))
			}

			os.Setenv(key, string(plaintext))
		}
	}

	return errs
}

func parseMasterKey(asset *Asset) ([]byte, error) {
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
		key, err = ioutil.ReadFile(asset.Layout()["config"] + "/" + env + ".key")
		if err != nil {
			return nil, ErrReadMasterKeyFile
		}
	}

	key = []byte(strings.Trim(string(key), "\n"))
	key = []byte(strings.Trim(string(key), " "))

	if len(key) == 0 {
		return nil, ErrMissingMasterKey
	}

	return key, nil
}
