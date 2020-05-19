package support

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config defines the application settings.
type Config struct {
	// AppyEnv indicates the environment that the codebase is running on and it
	// determines which config to use. By default, it is is "development" and
	// its corresponding config is "configs/.env.development".
	//
	// Note: APPY_ENV=test is used for unit tests.
	AppyEnv string `env:"APPY_ENV" envDefault:"development"`

	// AssetHost indicates the asset host to use with "assetPath()" for the
	// server-side rendering which is very useful it comes to hosting the static
	// assets on CDN. By default, it is "" which uses the current server host.
	AssetHost string `env:"ASSET_HOST" envDefault:""`

	// GQLPlaygroundEnabled indicates if the GraphQL playground is enabled. By
	// default, it is false.
	GQLPlaygroundEnabled bool `env:"GQL_PLAYGROUND_ENABLED" envDefault:"false"`

	// GQLPlaygroundPath indicates the GraphQL playground path to host at. By
	// default, it is "/docs/graphql".
	GQLPlaygroundPath string `env:"GQL_PLAYGROUND_PATH" envDefault:"/docs/graphql"`

	// GQLAPQCacheSize indicates how many APQ to persist in the memory at one
	// time. By default, it is 100. For more details about APQ, please refer to
	// https://gqlgen.com/reference/apq.
	GQLAPQCacheSize int `env:"GQL_APQ_CACHE_SIZE" envDefault:"100"`

	// GQLQueryCacheSize indicates how many queries to cache in the memory. By
	// default, it is 1000.
	GQLQueryCacheSize int `env:"GQL_QUERY_CACHE_SIZE" envDefault:"1000"`

	// GQLComplexityLimit indicates the query complexity which can be used to
	// mitigate DDoS attacks risk. By default, it is 1000.
	GQLComplexityLimit int `env:"GQL_COMPLEXITY_LIMIT" envDefault:"1000"`

	// GQLMultipartMaxMemory indicates the maximum number of bytes used to parse
	// a request body as multipart/form-data in memory, with the remainder stored
	// on disk in temporary files. By default, it is 0 (no limit).
	GQLMultipartMaxMemory int64 `env:"GQL_MULTIPART_MAX_MEMORY" envDefault:"0"`

	// GQLMultipartMaxUploadSize indicates the maximum number of bytes used to
	// parse a request body as multipart/form-data. By default, it is 0 (no limit).
	GQLMultipartMaxUploadSize int64 `env:"GQL_MULTIPART_MAX_UPLOAD_SIZE" envDefault:"0"`

	// GQLWebsocketKeepAliveDuration indicates how long the websocket connection
	// should be kept alive for sending subsequent messages without re-establishing
	// the connection which is an overhead. By default, it is 10s.
	GQLWebsocketKeepAliveDuration time.Duration `env:"GQL_WEBSOCKET_KEEP_ALIVE_DURATION" envDefault:"10s"`

	// HTTPGzipCompressLevel indicates the compression level used to compress the
	// HTTP response. By default, it is -1.
	//
	// Available options:
	// 	 - Default Compression = -1
	//   - No Compression      = 0
	//   - Fastest Compression = 1
	//   - Best Compression    = 9
	HTTPGzipCompressLevel int `env:"HTTP_GZIP_COMPRESS_LEVEL" envDefault:"-1"`

	// HTTPGzipExcludedExts indicates which file extensions not to compress. By
	// default, it is "".
	HTTPGzipExcludedExts []string `env:"HTTP_GZIP_EXCLUDED_EXTS" envDefault:""`

	// HTTPGzipExcludedPaths indicates which paths not to compress. By default,
	// it is "".
	HTTPGzipExcludedPaths []string `env:"HTTP_GZIP_EXCLUDED_PATHS" envDefault:""`

	// HTTPLogFilterParameters indicates which query parameters in the URL to
	// filter so that the sensitive information like password are masked in the
	// HTTP request log. By default, it is "password".
	HTTPLogFilterParameters []string `env:"HTTP_LOG_FILTER_PARAMETERS" envDefault:"password"`

	// HTTPHealthCheckPath indicates the path to check if the HTTP server is healthy.
	// This endpoint is a middleware that is designed to avoid redundant computing
	// resource usage. By default, it is "/health_check".
	//
	// In general, if your server is running behind a load balancer, this endpoint
	// will be served to inform the load balancer that the server is healthy and
	// ready to receive HTTP requests.
	HTTPHealthCheckPath string `env:"HTTP_HEALTH_CHECK_PATH" envDefault:"/health_check"`

	// HTTPHost indicates which host the HTTP server should be hosted at. By
	// default, it is "localhost". If you would like to connect to the HTTP server
	// from within your LAN network, use "0.0.0.0" instead.
	HTTPHost string `env:"HTTP_HOST" envDefault:"localhost"`

	// HTTPPort indicates which port the HTTP server should be hosted at. By
	// default, it is "3000".
	HTTPPort string `env:"HTTP_PORT" envDefault:"3000"`

	// HTTPGracefulShutdownTimeout indicates how long to wait for the HTTP server
	// to shut down so that any active connection is not interrupted by
	// SIGTERM/SIGINT. By default, it is "30s".
	HTTPGracefulShutdownTimeout time.Duration `env:"HTTP_GRACEFUL_SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// HTTPIdleTimeout is the maximum amount of time to wait for the next request
	// when keep-alives are enabled. If HTTPIdleTimeout is zero, the value of
	// HTTPReadTimeout is used. If both are zero, there is no timeout. By default,
	// it is "75s".
	HTTPIdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"75s"`

	// HTTPMaxHeaderBytes controls the maximum number of bytes the server will read
	// parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body. If zero,
	// http.DefaultMaxHeaderBytes (1 << 20 which is 1 MB) is used.
	HTTPMaxHeaderBytes int `env:"HTTP_MAX_HEADER_BYTES" envDefault:"0"`

	// HTTPReadTimeout is the maximum duration for reading the entire request,
	// including the body. Because HTTPReadTimeout does not let Handlers make
	// per-request decisions on each request body's acceptable deadline or upload
	// rate, most users will prefer to use HTTPReadHeaderTimeout. It is valid
	// to use them both. By default, it is "60s".
	HTTPReadTimeout time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"60s"`

	// HTTPReadHeaderTimeout is the amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and the
	// Handler can decide what is considered too slow for the body. If
	// HTTPReadHeaderTimeout is zero, the value of HTTPReadTimeout is used. If
	// both are zero, there is no timeout. By default, it is "60s".
	HTTPReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"60s"`

	// HTTPWriteTimeout is the maximum duration before timing out writes of the
	// response. It is reset whenever a new request's header is read. Like
	// HTTPReadTimeout, it does not let Handlers make decisions on a per-request
	// basis. By default, it is "60s".
	HTTPWriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"60s"`

	// HTTPSSLCertPath indicates which path to store the locally trusted SSL
	// certificates which are created using "go run . ssl:setup" command. By
	// default, it is "./tmp/ssl".
	HTTPSSLCertPath string `env:"HTTP_SSL_CERT_PATH" envDefault:"./tmp/ssl"`

	// HTTPSSLEnabled indicates if the HTTPS server should be enabled. When
	// enabled, please ensure the SSL certificates are created in the
	// "HTTPSSLCertPath". By default, it is false.
	HTTPSSLEnabled bool `env:"HTTP_SSL_ENABLED" envDefault:"false"`

	// HTTPSSLPort indicates which port the HTTPS server should be hosted at.
	// By default, it is "3443".
	HTTPSSLPort string `env:"HTTP_SSL_PORT" envDefault:"3443"`

	// HTTPSessionProvider indicates which store to use for storing the session
	// information. By default, it is "cookie".
	//
	// Available options:
	//   - cookie
	//   - redis
	HTTPSessionProvider string `env:"HTTP_SESSION_PROVIDER" envDefault:"cookie"`

	// HTTPSessionExpiration indicates how long till the session should expire.
	// By default, it is 1209600 in seconds.
	//
	// Note: If the HTTPSessionProvider is "redis", the same value will be used
	// to expire the session in Redis.
	HTTPSessionExpiration int `env:"HTTP_SESSION_EXPIRATION" envDefault:"1209600"`

	// HTTPSessionSecrets indicates the secrets to encrypt the session information.
	// By default, it is "".
	//
	// Note: Multiple values are accepted with comma delimited to allow easier
	// secret rotation.
	HTTPSessionSecrets [][]byte `env:"HTTP_SESSION_SECRETS,required" envDefault:""`

	// HTTPSessionRedisAddr indicates the Redis server to store the session. The
	// Redis session store will store the session's information and use cookie to
	// store the session ID. By default, it is "localhost:6379".
	//
	// Note: Please ensure that HTTPSessionProvider is configured to be "redis"
	// when using this.
	HTTPSessionRedisAddr string `env:"HTTP_SESSION_REDIS_ADDR" envDefault:"localhost:6379"`

	// HTTPSessionRedisPassword indicates the password to authenticate with the Redis
	// server. By default, it is "".
	HTTPSessionRedisPassword string `env:"HTTP_SESSION_REDIS_PASSWORD" envDefault:""`

	// HTTPSessionRedisDB indicates the Redis database to use. By default, it is "0".
	HTTPSessionRedisDB int `env:"HTTP_SESSION_REDIS_DB" envDefault:"0"`

	// HTTPSessionRedisPoolSize indicates how many maximum connections a CPU should
	// keep in its connections pool. By default, it is 10.
	HTTPSessionRedisPoolSize int `env:"HTTP_SESSION_REDIS_POOL_SIZE" envDefault:"10"`

	// HTTPSessionRedisPoolTimeout indicates how long to wait for a connection to
	// be returned from the connection pool before returning an error. By default,
	// it is "4s".
	HTTPSessionRedisPoolTimeout time.Duration `env:"HTTP_SESSION_REDIS_POOL_TIMEOUT" envDefault:"4s"`

	// HTTPSessionRedisMaxConnAge indicates how long a connection should be
	// kept alive. If it is 0, no connections will be closed based on the age.
	// By default, it is 0.
	HTTPSessionRedisMaxConnAge time.Duration `env:"HTTP_SESSION_REDIS_MAX_CONN_AGE" envDefault:"0"`

	// HTTPSessionRedisMinIdleConns indicates how many minimum idle connections
	// should remain in the connection pool at one time. If it is 0, there won't
	// be minimum idle connections remained in the connection pool. By default,
	// it is 0.
	HTTPSessionRedisMinIdleConns int `env:"HTTP_SESSION_REDIS_MIN_IDLE_CONNS" envDefault:"0"`

	// HTTPSessionRedisIdleCheckFrequency indicates how frequent the reaper checks
	// and closes the idle connections. If it is -1, the idle connections won't be
	// closed by the reaper. However, it will still be closed by the client if
	// HTTPSessionRedisIdleTimeout is set. By default, it is "1m".
	HTTPSessionRedisIdleCheckFrequency time.Duration `env:"HTTP_SESSION_REDIS_IDLE_CHECK_FREQUENCY" envDefault:"1m"`

	// HTTPSessionRedisIdleTimeout indicates how long the idle connection should
	// be kept around. If it is 0, the idle connections won't be closed. By
	// default, it is "5m".
	HTTPSessionRedisIdleTimeout time.Duration `env:"HTTP_SESSION_REDIS_IDLE_TIMEOUT" envDefault:"5m"`

	// HTTPSessionCookieName indicates the cookie name to use to store the session.
	// By default, it is "_session".
	HTTPSessionCookieName string `env:"HTTP_SESSION_COOKIE_NAME" envDefault:"_session"`

	// HTTPSessionCookieDomain indicates which domain the session cookie can be sent to.
	// By default, it is "localhost".
	HTTPSessionCookieDomain string `env:"HTTP_SESSION_COOKIE_DOMAIN" envDefault:"localhost"`

	// HTTPSessionCookieHTTPOnly indicates if the session cookie is only accessible
	// via HTTP request and not Javascript `Document.cookie` API.
	// By default, it is true.
	HTTPSessionCookieHTTPOnly bool `env:"HTTP_SESSION_COOKIE_HTTP_ONLY" envDefault:"true"`

	// HTTPSessionCookiePath indicates which URL path the session cookie can be
	// sent to. By default, it is "/".
	HTTPSessionCookiePath string `env:"HTTP_SESSION_COOKIE_PATH" envDefault:"/"`

	// HTTPSessionCookieSameSite indicates if the session cookie can be sent with
	// cross-site requests. By default, it is 1.
	//
	// Available options:
	//   - SameSiteDefaultMode = 1
	//   - SameSiteLaxMode     = 2
	//   - SameSiteStrictMode  = 3
	//   - SameSiteNoneMode    = 4
	HTTPSessionCookieSameSite http.SameSite `env:"HTTP_SESSION_COOKIE_SAME_SITE" envDefault:"1"`

	// HTTPSessionCookieSecure indicates if the session cookie can only be sent
	// with HTTPS request. By default, it is false.
	HTTPSessionCookieSecure bool `env:"HTTP_SESSION_COOKIE_SECURE" envDefault:"false"`

	// HTTPAllowedHosts indicates a list of fully qualified domain names that are
	// allowed to be processed by the HTTP server. By default, it is "" which
	// allows all domain names.
	HTTPAllowedHosts []string `env:"HTTP_ALLOWED_HOSTS" envDefault:""`

	// HTTPCSRFCookieDomain indicates which domain the CSRF cookie can be sent
	// to. By default, it is "localhost".
	HTTPCSRFCookieDomain string `env:"HTTP_CSRF_COOKIE_DOMAIN" envDefault:"localhost"`

	// HTTPCSRFCookieName indicates the cookie name to use to store the CSRF
	// token. By default, it is "_csrf_token".
	HTTPCSRFCookieName string `env:"HTTP_CSRF_COOKIE_NAME" envDefault:"_csrf_token"`

	// HTTPCSRFCookieHTTPOnly indicates if the CSRF cookie is only accessible via
	// HTTP request and not Javascript `Document.cookie` API. By default, it is true.
	HTTPCSRFCookieHTTPOnly bool `env:"HTTP_CSRF_COOKIE_HTTP_ONLY" envDefault:"true"`

	// HTTPCSRFCookieMaxAge indicates how long till the CSRF cookie should expire.
	// By default, it is 0 which is no expiration.
	HTTPCSRFCookieMaxAge int `env:"HTTP_CSRF_COOKIE_MAX_AGE" envDefault:"0"`

	// HTTPCSRFCookiePath indicates which URL path the CSRF cookie can be sent to.
	// By default, it is "/".
	HTTPCSRFCookiePath string `env:"HTTP_CSRF_COOKIE_PATH" envDefault:"/"`

	// HTTPCSRFCookieSameSite indicates if the CSRF cookie can be sent with
	// cross-site requests. By default, it is 1.
	//
	// Available options:
	//   - SameSiteDefaultMode = 1
	//   - SameSiteLaxMode     = 2
	//   - SameSiteStrictMode  = 3
	//   - SameSiteNoneMode    = 4
	HTTPCSRFCookieSameSite http.SameSite `env:"HTTP_CSRF_COOKIE_SAME_SITE" envDefault:"1"`

	// HTTPCSRFCookieSecure indicates if the session cookie can only be sent
	// with HTTPS request. By default, it is false.
	HTTPCSRFCookieSecure bool `env:"HTTP_CSRF_COOKIE_SECURE" envDefault:"false"`

	// HTTPCSRFAuthenticityFieldName indicates the POST form field name that
	// contains the authenticity token for CSRF check. By default, it is
	// "authenticity_token".
	HTTPCSRFAuthenticityFieldName string `env:"HTTP_CSRF_AUTHENTICITY_FIELD_NAME" envDefault:"authenticity_token"`

	// HTTPCSRFRequestHeader indicates the HTTP header that contains the
	// authenticity token for CSRF check. By default, it is "X-CSRF-Token".
	HTTPCSRFRequestHeader string `env:"HTTP_CSRF_REQUEST_HEADER" envDefault:"X-CSRF-Token"`

	// HTTPCSRFSecret indicates the secret to encrypt the CSRF cookie. By
	// default, it is "".
	HTTPCSRFSecret []byte `env:"HTTP_CSRF_SECRET,required" envDefault:""`

	// HTTPSSLRedirect indicates if the HTTP server should automatically
	// redirect HTTP requests to HTTPS. By default, it is false.
	HTTPSSLRedirect bool `env:"HTTP_SSL_REDIRECT" envDefault:"false"`

	// HTTPSSLTemporaryRedirect indicates if the HTTPSSLRedirect should
	// use 302 temporary redirect. By default, it is false which uses 301
	// permanent redirect.
	HTTPSSLTemporaryRedirect bool `env:"HTTP_SSL_TEMPORARY_REDIRECT" envDefault:"false"`

	// HTTPSSLHost indicates the host name to redirect to when redirecting HTTP
	// requests to HTTPS. By default, it is "" which uses the same host.
	HTTPSSLHost string `env:"HTTP_SSL_HOST" envDefault:""`

	// HTTPSSLProxyHeaders indicates the SSL proxy headers to verify when the
	// server is running behind a proxy like Heroku to ensure the request comes
	// through the SSL proxy server. By default, it is "X-Forwarded-Proto:https".
	HTTPSSLProxyHeaders map[string]string `env:"HTTP_SSL_PROXY_HEADERS" envDefault:"X-Forwarded-Proto:https"`

	// HTTPSTSSeconds indicates the max-age of the "Strict-Transport-Security"
	// response header. By default, it is 0 which would NOT include the header.
	HTTPSTSSeconds int64 `env:"HTTP_STS_SECONDS" envDefault:"0"`

	// HTTPSTSIncludeSubdomains indicates if "includeSubdomains" should be
	// appended to the "Strict-Transport-Security" response header. By default,
	// it is false.
	HTTPSTSIncludeSubdomains bool `env:"HTTP_STS_INCLUDE_SUBDOMAINS" envDefault:"false"`

	// HTTPFrameDeny indicates if the page cannot be contained in an iframe by
	// setting the "X-Frame-Options: DENY" response header. By default, it is false.
	HTTPFrameDeny bool `env:"HTTP_FRAME_DENY" envDefault:"false"`

	// HTTPCustomFrameOptionsValue indicates the custom value to be set for the
	// "X-Frame-Options" response header. By default, it is "" which doesn't
	// override the "DENY" value.
	HTTPCustomFrameOptionsValue string `env:"HTTP_CUSTOM_FRAME_OPTIONS_VALUE" envDefault:""`

	// HTTPContentTypeNosniff indicates if the "X-Content-Type-Options: nosniff"
	// response header should be set. By default, it is false.
	HTTPContentTypeNosniff bool `env:"HTTP_CONTENT_TYPE_NOSNIFF" envDefault:"false"`

	// HTTPBrowserXSSFilter indicates if the "X-XSS-Protection: 1; mode=block"
	// response header should be set. By default, it is false.
	HTTPBrowserXSSFilter bool `env:"HTTP_BROWSER_XSS_FILTER" envDefault:"false"`

	// HTTPContentSecurityPolicy indicates the custom value to be set for
	// "Content-Security-Policy" response header. By default, it is "".
	HTTPContentSecurityPolicy string `env:"HTTP_CONTENT_SECURITY_POLICY" envDefault:""`

	// HTTPReferrerPolicy indicates the referrer information that is sent in the
	// "Referrer" header to be included with the requests made. By default, it is "".
	HTTPReferrerPolicy string `env:"HTTP_REFERRER_POLICY" envDefault:""`

	// HTTPIENoOpen indicates if it should prevent Internet Explorer from
	// executing downloads in your site’s context. By default, it is false.
	HTTPIENoOpen bool `env:"HTTP_IE_NO_OPEN" envDefault:"false"`

	// I18nDefaultLocale indicates the default locale to use for translations in
	// handlers/mailers/views when the desire locale is not found. By default, it is "en".
	//
	// Note: If the locale is "en", the translation file would be "pkg/locales/en.yml".
	I18nDefaultLocale string `env:"I18N_DEFAULT_LOCALE" envDefault:"en"`

	// MailerSMTPAddr indicates the SMTP server hostname that sends out email.
	// By default, it is "".
	MailerSMTPAddr string `env:"MAILER_SMTP_ADDR" envDefault:""`

	// MailerSMTPPlainAuthIdentity indicates the SMTP plain auth identity to use
	// for sending out email. By default, it is "".
	//
	// Note: This is normally not needed.
	MailerSMTPPlainAuthIdentity string `env:"MAILER_SMTP_PLAIN_AUTH_IDENTITY" envDefault:""`

	// MailerSMTPPlainAuthUsername indicates the SMTP plain auth username to use
	// for sending out email. By default, it is "".
	MailerSMTPPlainAuthUsername string `env:"MAILER_SMTP_PLAIN_AUTH_USERNAME" envDefault:""`

	// MailerSMTPPlainAuthPassword indicates the SMTP plain auth password to use
	// for sending out email. By default, it is "".
	MailerSMTPPlainAuthPassword string `env:"MAILER_SMTP_PLAIN_AUTH_PASSWORD" envDefault:""`

	// MailerSMTPPlainAuthHost indicates the SMTP plain auth server host to use for
	// sending out email. By default, it is "".
	MailerSMTPPlainAuthHost string `env:"MAILER_SMTP_PLAIN_AUTH_HOST" envDefault:""`

	// MailerPreviewPath indicates the path for previewing the mailers. By default,
	// it is "/appy/mailers".
	MailerPreviewPath string `env:"MAILER_PREVIEW_PATH" envDefault:"/appy/mailers"`

	// WorkerRedisSentinelAddrs indicates the Redis sentinel hosts to connect to.
	// By default, it is "".
	//
	// Note: If this is configured to non-empty string, both WorkerRedisAddr or
	// WorkerRedisURL will be ignored.
	WorkerRedisSentinelAddrs []string `env:"WORKER_REDIS_SENTINEL_ADDRS" envDefault:""`

	// WorkerRedisSentinelDB indicates the Redis DB to connect in the sentinel hosts.
	// By default, it is 0.
	WorkerRedisSentinelDB int `env:"WORKER_REDIS_SENTINEL_DB" envDefault:"0"`

	// WorkerRedisSentinelMasterName indicates the Redis sentinel master name to
	// connect to. By default, it is "".
	WorkerRedisSentinelMasterName string `env:"WORKER_REDIS_SENTINEL_MASTER_NAME" envDefault:""`

	// WorkerRedisSentinelPassword indicates the password used to connect to the
	// sentinel hosts. By default, it is "".
	WorkerRedisSentinelPassword string `env:"WORKER_REDIS_SENTINEL_PASSWORD" envDefault:""`

	// WorkerRedisSentinelPoolSize indicates the connection pool size for the
	// sentinel hosts. By default, it is 25.
	WorkerRedisSentinelPoolSize int `env:"WORKER_REDIS_SENTINEL_POOL_SIZE" envDefault:"25"`

	// WorkerRedisAddr indicates the Redis hostname to connect. By default, it is
	// "localhost:6379".
	WorkerRedisAddr string `env:"WORKER_REDIS_ADDR" envDefault:"localhost:6379"`

	// WorkerRedisDB indicates the Redis DB to connect. By default, it is 0.
	WorkerRedisDB int `env:"WORKER_REDIS_DB" envDefault:"0"`

	// WorkerRedisPassword indicates the password used to connect to the Redis
	// server. By default, it is "".
	WorkerRedisPassword string `env:"WORKER_REDIS_PASSWORD" envDefault:""`

	// WorkerRedisPoolSize indicates the connection pool size for the Redis
	// server. By default, it is 25.
	WorkerRedisPoolSize int `env:"WORKER_REDIS_POOL_SIZE" envDefault:"25"`

	// WorkerRedisURL indicates the Redis URL to connect to. By default, it is "".
	WorkerRedisURL string `env:"WORKER_REDIS_URL" envDefault:""`

	// WorkerConcurrency indicates how many background jobs should be processed
	// at one time. By default, it is 25.
	WorkerConcurrency int `env:"WORKER_CONCURRENCY" envDefault:"25"`

	// WorkerQueues indicates how many queues to process and the number followed
	// is the priority. By default, it is "default:10".
	//
	// If the value is "critical:6,default:3,low:1", this will allow the worker
	// to process 3 queues as below:
	// - tasks in critical queue will be processed 60% of the time
	// - tasks in default queue will be processed 30% of the time
	// - tasks in low queue will be processed 10% of the time
	WorkerQueues map[string]int `env:"WORKER_QUEUES" envDefault:"default:10"`

	// WorkerStrictPriority indicates if the worker should strictly follow the
	// priority to process the background jobs. By default, it is false.
	//
	// If the value is true, the queues with higher priority is always processed
	// first, and queues with lower priority is processed only if all the other
	// queues with higher priorities are empty.
	WorkerStrictPriority bool `env:"WORKER_STRICT_PRIORITY" envDefault:"false"`

	// WorkerGracefulShutdownTimeout indicates how long to wait for the worker
	// to shut down so that any active job processing is not interrupted by
	// SIGTERM/SIGINT. By default, it is "30s".
	WorkerGracefulShutdownTimeout time.Duration `env:"WORKER_GRACEFUL_SHUTDOWN_TIMEOUT" envDefault:"30s"`

	asset     AssetManager
	errors    []error
	masterKey []byte
}

// NewConfig initializes Config instance.
func NewConfig(asset AssetManager, logger *Logger) *Config {
	config := &Config{
		asset:  asset,
		errors: []error{},
	}

	masterKey, err := parseMasterKey(asset)
	if err != nil {
		config.errors = append(config.errors, err)
	}

	if masterKey != nil {
		config.masterKey = masterKey

		if errs := config.decrypt(asset); len(errs) > 0 {
			config.errors = append(config.errors, errs...)
		}

		if err := ParseEnv(config); err != nil {
			config.errors = append(config.errors, err)
		}
	}

	return config
}

// Errors returns all the config retrieving/parsing errors.
func (c *Config) Errors() []error {
	return c.errors
}

// IsProtectedEnv is used to protect the app from being destroyed by a command
// accidentally.
func (c *Config) IsProtectedEnv() bool {
	return c.AppyEnv == "production"
}

// MasterKey returns the master key for the current environment.
func (c *Config) MasterKey() []byte {
	return c.masterKey
}

// Path returns the config path.
func (c *Config) Path() string {
	return c.asset.Layout().config + "/.env." + os.Getenv("APPY_ENV")
}

func (c *Config) decrypt(asset AssetManager) []error {
	reader, err := asset.Open(c.Path())
	if err != nil {
		return []error{err}
	}

	envMap, err := godotenv.Parse(reader)
	if err != nil {
		os.Clearenv()
		return []error{err}
	}

	errs := []error{}
	if len(c.masterKey) != 0 {
		// Parse the environment variables that aren't defined in the .env config
		// file which can be used to override the ones defined in .env config file.
		origEnvMap := parseEnvVariables()

		for key, value := range envMap {
			if origEnvMap[key] || value == "" {
				continue
			}

			decoded, err := hex.DecodeString(value)
			if err != nil {
				errs = append(errs, err)
			}

			plaintext, err := AESDecrypt(decoded, c.MasterKey())
			if len(plaintext) < 1 || err != nil {
				errs = append(
					errs,
					fmt.Errorf("unable to decrypt '%s' value in '%s'", key, c.Path()),
				)
			}

			os.Setenv(key, string(plaintext))
		}
	}

	return errs
}

func parseMasterKey(asset AssetManager) ([]byte, error) {
	var (
		err       error
		masterKey []byte
	)

	env := "development"
	if os.Getenv("APPY_ENV") != "" {
		env = os.Getenv("APPY_ENV")
	}

	if os.Getenv("APPY_MASTER_KEY") != "" {
		masterKey = []byte(os.Getenv("APPY_MASTER_KEY"))
	}

	if len(masterKey) == 0 && IsDebugBuild() {
		masterKey, err = asset.ReadFile(asset.Layout().config + "/" + env + ".key")
		if err != nil {
			return nil, ErrReadMasterKeyFile
		}
	}

	masterKey = []byte(strings.Trim(string(masterKey), "\n"))
	masterKey = []byte(strings.Trim(string(masterKey), " "))

	if len(masterKey) == 0 {
		return nil, ErrMissingMasterKey
	}

	return masterKey, nil
}

func parseEnvVariables() map[string]bool {
	envMap := map[string]bool{}
	envs := os.Environ()
	for _, env := range envs {
		key := strings.Split(env, "=")[0]
		envMap[key] = true
	}

	return envMap
}
