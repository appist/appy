package appy

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type (
	// Config provides the configuration functionality.
	Config struct {
		AppyEnv string `env:"APPY_ENV" envDefault:"development"`

		// GraphQL related configuration.
		GQLPlaygroundEnabled          bool          `env:"GQL_PLAYGROUND_ENABLED" envDefault:"false"`
		GQLPlaygroundPath             string        `env:"GQL_PLAYGROUND_PATH" envDefault:"/docs/graphql"`
		GQLCacheSize                  int           `env:"GQL_CACHE_SIZE" envDefault:"1000"`
		GQLComplexityLimit            int           `env:"GQL_COMPLEXITY_LIMIT" envDefault:"200"`
		GQLUploadMaxMemory            int64         `env:"GQL_UPLOAD_MAX_MEMORY" envDefault:"100000000"`
		GQLUploadMaxSize              int64         `env:"GQL_UPLOAD_MAX_SIZE" envDefault:"100000000"`
		GQLWebsocketKeepAliveDuration time.Duration `env:"GQL_WEBSOCKET_KEEP_ALIVE_DURATION" envDefault:"25s"`

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
		HTTPSessionCookieSecure   bool   `env:"HTTP_SESSION_COOKIE_SECURE" envDefault:"false"`

		// Session related configuration using redis pool.
		HTTPSessionRedisAddr            string        `env:"HTTP_SESSION_REDIS_ADDR" envDefault:"localhost:6379"`
		HTTPSessionRedisAuth            string        `env:"HTTP_SESSION_REDIS_AUTH" envDefault:""`
		HTTPSessionRedisDb              string        `env:"HTTP_SESSION_REDIS_DB" envDefault:"0"`
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
		HTTPCSRFCookieMaxAge        int               `env:"HTTP_CSRF_COOKIE_MAX_AGE" envDefault:"0"`
		HTTPCSRFCookieName          string            `env:"HTTP_CSRF_COOKIE_NAME" envDefault:"_csrf_token"`
		HTTPCSRFCookiePath          string            `env:"HTTP_CSRF_COOKIE_PATH" envDefault:"/"`
		HTTPCSRFCookieSecure        bool              `env:"HTTP_CSRF_COOKIE_SECURE" envDefault:"false"`
		HTTPCSRFFieldName           string            `env:"HTTP_CSRF_FIELD_NAME" envDefault:"authenticity_token"`
		HTTPCSRFRequestHeader       string            `env:"HTTP_CSRF_REQUEST_HEADER" envDefault:"X-CSRF-Token"`
		HTTPCSRFSecret              []byte            `env:"HTTP_CSRF_SECRET,required" envDefault:""`
		HTTPSSLRedirect             bool              `env:"HTTP_SSL_REDIRECT" envDefault:"false"`
		HTTPSSLTemporaryRedirect    bool              `env:"HTTP_SSL_TEMPORARY_REDIRECT" envDefault:"false"`
		HTTPSSLHost                 string            `env:"HTTP_SSL_HOST" envDefault:"localhost:3443"`
		HTTPSTSSeconds              int64             `env:"HTTP_STS_SECONDS" envDefault:"0"`
		HTTPSTSIncludeSubdomains    bool              `env:"HTTP_STS_INCLUDE_SUBDOMAINS" envDefault:"false"`
		HTTPFrameDeny               bool              `env:"HTTP_FRAME_DENY" envDefault:"true"`
		HTTPCustomFrameOptionsValue string            `env:"HTTP_CUSTOM_FRAME_OPTIONS_VALUE" envDefault:""`
		HTTPContentTypeNosniff      bool              `env:"HTTP_CONTENT_TYPE_NOSNIFF" envDefault:"false"`
		HTTPBrowserXSSFilter        bool              `env:"HTTP_BROWSER_XSS_FILTER" envDefault:"false"`
		HTTPContentSecurityPolicy   string            `env:"HTTP_CONTENT_SECURITY_POLICY" envDefault:""`
		HTTPReferrerPolicy          string            `env:"HTTP_REFERRER_POLICY" envDefault:""`
		HTTPIENoOpen                bool              `env:"HTTP_IE_NO_OPEN" envDefault:"false"`
		HTTPSSLProxyHeaders         map[string]string `env:"HTTP_SSL_PROXY_HEADERS" envDefault:""`

		build, path string
		masterKey   []byte
		errors      []error
	}
)

var (
	_csrPaths = map[string]string{
		"root": "web",
	}

	_ssrPaths = map[string]string{
		"root":   ".ssr",
		"docker": ".docker",
		"config": "pkg/config",
		"locale": "pkg/locales",
		"view":   "pkg/views",
	}

	_staticExtRegex = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
)

// NewConfig initializes Config instance.
func NewConfig(build string, logger *Logger, assets http.FileSystem) *Config {
	var (
		errs []error
	)

	masterKey, err := parseMasterKey()
	if err != nil {
		errs = append(errs, err)
	}

	config := &Config{}
	if masterKey != nil {
		config.path = configPath(build)
		config.masterKey = masterKey
		decryptErrs := decryptConfig(build, config.path, assets, masterKey)
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

// IsConfigErrored is used to check if config contains any error during initialization.
func IsConfigErrored(config *Config, logger *Logger) bool {
	if config != nil && config.errors != nil {
		for _, err := range config.errors {
			logger.Info(err.Error())
		}

		return true
	}

	return false
}

// IsProtectedEnv is used to protect the app from being destroyed by a command accidentally.
func IsProtectedEnv(config *Config) bool {
	if config.AppyEnv == "production" {
		return true
	}

	return false
}

func configPath(build string) string {
	path := _ssrPaths["config"] + "/.env." + os.Getenv("APPY_ENV")
	if build == DebugBuild {
		return path
	}

	return _ssrPaths["root"] + "/" + path
}

func decryptConfig(build, path string, assets http.FileSystem, masterKey []byte) []error {
	var (
		file io.Reader
		err  error
	)

	if build == DebugBuild {
		file, err = os.Open(path)
		if err != nil {
			return []error{err}
		}
	} else {
		if assets != nil {
			file, err = assets.Open(path)
			if err != nil {
				return []error{err}
			}
		}

		if file == nil {
			return []error{ErrNoConfigInAssets}
		}
	}

	envMap, err := godotenv.Parse(file)
	if err != nil {
		return []error{err}
	}

	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	var errs []error
	if len(masterKey) != 0 {
		for key, value := range envMap {
			if !currentEnv[key] {
				decodeStr, _ := hex.DecodeString(value)
				plaintext, err := AESDecrypt(decodeStr, masterKey)
				if len(plaintext) < 1 || err != nil {
					errs = append(errs, fmt.Errorf("unable to decrypt '%s' value in '%s'", key, path))
				}

				os.Setenv(key, string(plaintext))
			}
		}
	}

	return errs
}

func parseDbConfig() (map[string]DbConfig, []error) {
	var (
		err  error
		errs []error
	)
	dbConfig := map[string]DbConfig{}
	dbNames := []string{}

	for _, val := range os.Environ() {
		re := regexp.MustCompile("DB_ADDR_(.*)")
		match := re.FindStringSubmatch(val)

		if len(match) > 1 {
			splits := strings.Split(match[1], "=")
			dbNames = append(dbNames, splits[0])
		}
	}

	for _, dbName := range dbNames {
		schemaSearchPath := "public"
		if val, ok := os.LookupEnv("DB_SCHEMA_SEARCH_PATH_" + dbName); ok && val != "" {
			schemaSearchPath = val
		}

		addr := "0.0.0.0:5432"
		if val, ok := os.LookupEnv("DB_ADDR_" + dbName); ok && val != "" {
			addr = val
		}

		user := "postgres"
		if val, ok := os.LookupEnv("DB_USER_" + dbName); ok && val != "" {
			user = val
		}

		password := "postgres"
		if val, ok := os.LookupEnv("DB_PASSWORD_" + dbName); ok && val != "" {
			password = val
		}

		database := "appy"
		if val, ok := os.LookupEnv("DB_NAME_" + dbName); ok && val != "" {
			database = val
		}

		appName := "appy"
		if val, ok := os.LookupEnv("DB_APP_NAME_" + dbName); ok && val != "" {
			appName = val
		}

		replica := false
		if val, ok := os.LookupEnv("DB_REPLICA_" + dbName); ok && val != "" {
			replica, err = strconv.ParseBool(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		maxRetries := 0
		if val, ok := os.LookupEnv("DB_MAX_RETRIES_" + dbName); ok && val != "" {
			maxRetries, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		retryStatement := false
		if val, ok := os.LookupEnv("DB_RETRY_STATEMENT_" + dbName); ok && val != "" {
			retryStatement, err = strconv.ParseBool(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		poolSize := 10
		if val, ok := os.LookupEnv("DB_POOL_SIZE_" + dbName); ok && val != "" {
			poolSize, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		poolTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_POOL_TIMEOUT_" + dbName); ok && val != "" {
			poolTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		minIdleConns := 0
		if val, ok := os.LookupEnv("DB_MIN_IDLE_CONNS_" + dbName); ok && val != "" {
			minIdleConns, err = strconv.Atoi(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		maxConnAge := 0 * time.Second
		if val, ok := os.LookupEnv("DB_MAX_CONN_AGE_" + dbName); ok && val != "" {
			maxConnAge, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		dialTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_DIAL_TIMEOUT_" + dbName); ok && val != "" {
			dialTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		idleCheckFrequency := 1 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_CHECK_FREQUENCY_" + dbName); ok && val != "" {
			idleCheckFrequency, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		idleTimeout := 5 * time.Minute
		if val, ok := os.LookupEnv("DB_IDLE_TIMEOUT_" + dbName); ok && val != "" {
			idleTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		readTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_READ_TIMEOUT_" + dbName); ok && val != "" {
			readTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		writeTimeout := 30 * time.Second
		if val, ok := os.LookupEnv("DB_WRITE_TIMEOUT_" + dbName); ok && val != "" {
			writeTimeout, err = time.ParseDuration(val)
			if err != nil {
				errs = append(errs, err)
			}
		}

		schemaMigrationsTable := "schema_migrations"
		if val, ok := os.LookupEnv("DB_SCHEMA_MIGRATIONS_TABLE_" + dbName); ok && val != "" {
			schemaMigrationsTable = val
		}

		config := DbConfig{}
		config.ApplicationName = appName
		config.Addr = addr
		config.User = user
		config.Password = password
		config.Database = database
		config.MaxRetries = maxRetries
		config.PoolSize = poolSize
		config.PoolTimeout = poolTimeout
		config.DialTimeout = dialTimeout
		config.IdleCheckFrequency = idleCheckFrequency
		config.IdleTimeout = idleTimeout
		config.ReadTimeout = readTimeout
		config.WriteTimeout = writeTimeout
		config.RetryStatementTimeout = retryStatement
		config.MinIdleConns = minIdleConns
		config.MaxConnAge = maxConnAge
		config.Replica = replica
		config.SchemaSearchPath = schemaSearchPath
		config.SchemaMigrationsTable = schemaMigrationsTable
		config.OnConnect = func(conn *DbConn) error {
			_, err := conn.Exec("SET search_path=? /* appy framework */", schemaSearchPath)
			if err != nil {
				return err
			}

			return nil
		}

		dbConfig[ToCamelCase(dbName)] = config
	}

	return dbConfig, errs
}

func parseMasterKey() ([]byte, error) {
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

	if len(key) == 0 {
		if Build == DebugBuild {
			key, err = ioutil.ReadFile(_ssrPaths["config"] + "/" + env + ".key")
			if err != nil {
				return nil, ErrReadMasterKeyFile
			}
		}
	}

	key = []byte(strings.Trim(string(key), "\n"))
	key = []byte(strings.Trim(string(key), " "))

	if len(key) == 0 {
		return nil, ErrNoMasterKey
	}

	return key, nil
}
