package db

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.charczuk.com/sdk/configutil"
)

const (
	// DefaultEngine is the default database engine.
	DefaultEngine = "pgx" // "postgres"
)

const (
	// EnvVarDBEngine is the environment variable used to set the Go `sql` driver.
	EnvVarDBEngine = "DB_ENGINE"
	// EnvVarDatabaseURL is the environment variable used to set the entire
	// database connection string.
	EnvVarDatabaseURL = "DATABASE_URL"
	// EnvVarDBHost is the environment variable used to set the host in a
	// database connection string.
	EnvVarDBHost = "DB_HOST"
	// EnvVarDBPort is the environment variable used to set the port in a
	// database connection string.
	EnvVarDBPort = "DB_PORT"
	// EnvVarDBName is the environment variable used to set the database name
	// in a database connection string.
	EnvVarDBName = "DB_NAME"
	// EnvVarDBSchema is the environment variable used to set the database
	// schema in a database connection string.
	EnvVarDBSchema = "DB_SCHEMA"
	// EnvVarDBApplicationName is the environment variable used to set the
	// `application_name` configuration parameter in a `lib/pq` connection
	// string.
	//
	// See: https://www.postgresql.org/docs/12/runtime-config-logging.html#GUC-APPLICATION-NAME
	EnvVarDBApplicationName = "DB_APPLICATION_NAME"
	// EnvVarDBUser is the environment variable used to set the user in a
	// database connection string.
	EnvVarDBUser = "DB_USER"
	// EnvVarDBPassword is the environment variable used to set the password
	// in a database connection string.
	EnvVarDBPassword = "DB_PASSWORD"
	// EnvVarDBConnectTimeout is is the environment variable used to set the
	// connect timeout in a database connection string.
	EnvVarDBConnectTimeout = "DB_CONNECT_TIMEOUT"
	// EnvVarDBLockTimeout is is the environment variable used to set the lock
	// timeout on a database config.
	EnvVarDBLockTimeout = "DB_LOCK_TIMEOUT"
	// EnvVarDBStatementTimeout is is the environment variable used to set the
	// statement timeout on a database config.
	EnvVarDBStatementTimeout = "DB_STATEMENT_TIMEOUT"
	// EnvVarDBSSLMode is the environment variable used to set the SSL mode in
	// a database connection string.
	EnvVarDBSSLMode = "DB_SSLMODE"
	// EnvVarDBIdleConnections is the environment variable used to set the
	// maximum number of idle connections allowed in a connection pool.
	EnvVarDBIdleConnections = "DB_IDLE_CONNECTIONS"
	// EnvVarDBMaxConnections is the environment variable used to set the
	// maximum number of connections allowed in a connection pool.
	EnvVarDBMaxConnections = "DB_MAX_CONNECTIONS"
	// EnvVarDBMaxLifetime is the environment variable used to set the maximum
	// lifetime of a connection in a connection pool.
	EnvVarDBMaxLifetime = "DB_MAX_LIFETIME"
	// EnvVarDBMaxIdleTime is the environment variable used to set the maximum
	// time a connection can be idle.
	EnvVarDBMaxIdleTime = "DB_MAX_IDLE_TIME"
	// EnvVarDBBufferPoolSize is the environment variable used to set the buffer
	// pool size on a connection in a connection pool.
	EnvVarDBBufferPoolSize = "DB_BUFFER_POOL_SIZE"
	// EnvVarDBDialect is the environment variable used to set the dialect
	// on a connection configuration (e.g. `postgres` or `cockroachdb`).
	EnvVarDBDialect = "DB_DIALECT"
)

const (
	// DefaultHost is the default database hostname, typically used
	// when developing locally.
	DefaultHost = "localhost"
	// DefaultPort is the default postgres port.
	DefaultPort = "5432"
	// DefaultDatabase is the default database to connect to, we use
	// `postgres` to not pollute the template databases.
	DefaultDatabase = "postgres"

	// DefaultSchema is the default schema to connect to
	DefaultSchema = "public"
)

const (
	// SSLModeDisable is an ssl mode.
	// Postgres Docs: "I don't care about security, and I don't want to pay the overhead of encryption."
	SSLModeDisable = "disable"
	// SSLModeAllow is an ssl mode.
	// Postgres Docs: "I don't care about security, but I will pay the overhead of encryption if the server insists on it."
	SSLModeAllow = "allow"
	// SSLModePrefer is an ssl mode.
	// Postgres Docs: "I don't care about encryption, but I wish to pay the overhead of encryption if the server supports it"
	SSLModePrefer = "prefer"
	// SSLModeRequire is an ssl mode.
	// Postgres Docs: "I want my data to be encrypted, and I accept the overhead. I trust that the network will make sure I always connect to the server I want."
	SSLModeRequire = "require"
	// SSLModeVerifyCA is an ssl mode.
	// Postgres Docs: "I want my data encrypted, and I accept the overhead. I want to be sure that I connect to a server that I trust."
	SSLModeVerifyCA = "verify-ca"
	// SSLModeVerifyFull is an ssl mode.
	// Postgres Docs: "I want my data encrypted, and I accept the overhead. I want to be sure that I connect to a server I trust, and that it's the one I specify."
	SSLModeVerifyFull = "verify-full"
)

const (
	// DefaultConnectTimeout is the default connect timeout.
	DefaultConnectTimeout = 5 * time.Second
	// DefaultIdleConnections is the default number of idle connections.
	DefaultIdleConnections = 16
	// DefaultMaxConnections is the default maximum number of connections.
	DefaultMaxConnections = 32
	// DefaultMaxLifetime is the default maximum lifetime of driver connections.
	DefaultMaxLifetime = time.Duration(0)
	// DefaultMaxIdleTime is the default maximum idle time of driver connections.
	DefaultMaxIdleTime = time.Duration(0)
	// DefaultBufferPoolSize is the default number of buffer pool entries to maintain.
	DefaultBufferPoolSize = 1024
)

// Config is a set of connection config options.
type Config struct {
	// Engine is the database engine.
	Engine string `json:"engine,omitempty" yaml:"engine,omitempty"`
	// DSN is a fully formed DSN (this skips DSN formation from all other variables outside `schema`).
	DSN string `json:"dsn,omitempty" yaml:"dsn,omitempty"`
	// Host is the server to connect to.
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	// Port is the port to connect to.
	Port string `json:"port,omitempty" yaml:"port,omitempty"`
	// DBName is the database name
	Database string `json:"database,omitempty" yaml:"database,omitempty"`
	// Schema is the application schema within the database, defaults to `public`. This schema is used to set the
	// Postgres "search_path" If you want to reference tables in other schemas, you'll need to specify those schemas
	// in your queries e.g. "SELECT * FROM schema_two.table_one..."
	// Using the public schema in a production application is considered bad practice as newly created roles will have
	// visibility into this data by default. We strongly recommend specifying this option and using a schema that is
	// owned by your service's role
	// We recommend against setting a multi-schema search_path, but if you really want to, you provide multiple comma-
	// separated schema names as the value for this config, or you can dbc.Invoke().Exec a SET statement on a newly
	// opened connection such as "SET search_path = 'schema_one,schema_two';" Again, we recommend against this practice
	// and encourage you to specify schema names beyond the first in your queries.
	Schema string `json:"schema,omitempty" yaml:"schema,omitempty"`
	// ApplicationName is the name set by an application connection to a database
	// server, intended to be transmitted in the connection string. It can be
	// used to uniquely identify an application and will be included in the
	// `pg_stat_activity` view.
	//
	// See: https://www.postgresql.org/docs/12/runtime-config-logging.html#GUC-APPLICATION-NAME
	ApplicationName string `json:"applicationName,omitempty" yaml:"applicationName,omitempty"`
	// Username is the username for the connection via password auth.
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	// Password is the password for the connection via password auth.
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	// ConnectTimeout determines the maximum wait for connection. The minimum
	// allowed timeout is 2 seconds, so anything below is treated the same
	// as unset. PostgreSQL will only accept second precision so this value will be
	// rounded to the nearest second before being set on a connection string.
	// Use `Validate()` to confirm that `ConnectTimeout` is exact to second
	// precision.
	//
	// See: https://www.postgresql.org/docs/10/libpq-connect.html#LIBPQ-CONNECT-CONNECT-TIMEOUT
	ConnectTimeout time.Duration `json:"connectTimeout,omitempty" yaml:"connectTimeout,omitempty"`
	// LockTimeout is the timeout to use when attempting to acquire a lock.
	// PostgreSQL will only accept millisecond precision so this value will be
	// rounded to the nearest millisecond before being set on a connection string.
	// Use `Validate()` to confirm that `LockTimeout` is exact to millisecond
	// precision.
	//
	// See: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
	LockTimeout time.Duration `json:"lockTimeout,omitempty" yaml:"lockTimeout,omitempty"`
	// StatementTimeout is the timeout to use when invoking a SQL statement.
	// PostgreSQL will only accept millisecond precision so this value will be
	// rounded to the nearest millisecond before being set on a connection string.
	// Use `Validate()` to confirm that `StatementTimeout` is exact to millisecond
	// precision.
	//
	// See: https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
	StatementTimeout time.Duration `json:"statementTimeout,omitempty" yaml:"statementTimeout,omitempty"`
	// SSLMode is the sslmode for the connection.
	SSLMode string `json:"sslMode,omitempty" yaml:"sslMode,omitempty"`
	// IdleConnections is the number of idle connections.
	IdleConnections int `json:"idleConnections,omitempty" yaml:"idleConnections,omitempty"`
	// MaxConnections is the maximum number of connections.
	MaxConnections int `json:"maxConnections,omitempty" yaml:"maxConnections,omitempty"`
	// MaxLifetime is the maximum time a connection can be open.
	MaxLifetime time.Duration `json:"maxLifetime,omitempty" yaml:"maxLifetime,omitempty"`
	// MaxIdleTime is the maximum time a connection can be idle.
	MaxIdleTime time.Duration `json:"maxIdleTime,omitempty" yaml:"maxIdleTime,omitempty"`
	// BufferPoolSize is the number of query composition buffers to maintain.
	BufferPoolSize int `json:"bufferPoolSize,omitempty" yaml:"bufferPoolSize,omitempty"`
	// Dialect includes hints to tweak specific sql semantics by database connection.
	Dialect string `json:"dialect,omitempty" yaml:"dialect,omitempty"`
}

// IsZero returns if the config is unset.
func (c Config) IsZero() bool {
	return c.DSN == "" && c.Host == "" && c.Port == "" && c.Database == "" && c.Schema == "" && c.Username == "" && c.Password == "" && c.SSLMode == ""
}

// Resolve applies any external data sources to the config.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.Engine, configutil.Env[string](EnvVarDBEngine), configutil.Lazy(&c.Engine), configutil.Const(DefaultEngine)),
		configutil.Set(&c.DSN, configutil.Env[string](EnvVarDatabaseURL), configutil.Lazy(&c.DSN)),
		configutil.Set(&c.Host, configutil.Env[string](EnvVarDBHost), configutil.Lazy(&c.Host), configutil.Const(DefaultHost)),
		configutil.Set(&c.Port, configutil.Env[string](EnvVarDBPort), configutil.Lazy(&c.Port), configutil.Const(DefaultPort)),
		configutil.Set(&c.Database, configutil.Env[string](EnvVarDBName), configutil.Lazy(&c.Database), configutil.Const(DefaultDatabase)),
		configutil.Set(&c.Schema, configutil.Env[string](EnvVarDBSchema), configutil.Lazy(&c.Schema)),
		configutil.Set(&c.ApplicationName, configutil.Env[string](EnvVarDBApplicationName), configutil.Const(c.ApplicationName)),
		configutil.Set(&c.Username, configutil.Env[string](EnvVarDBUser), configutil.Lazy(&c.Username), configutil.Env[string]("USER")),
		configutil.Set(&c.Password, configutil.Env[string](EnvVarDBPassword), configutil.Lazy(&c.Password)),
		configutil.Set(&c.ConnectTimeout, configutil.Env[time.Duration](EnvVarDBConnectTimeout), configutil.Lazy(&c.ConnectTimeout), configutil.Const(DefaultConnectTimeout)),
		configutil.Set(&c.LockTimeout, configutil.Env[time.Duration](EnvVarDBLockTimeout), configutil.Lazy(&c.LockTimeout)),
		configutil.Set(&c.StatementTimeout, configutil.Env[time.Duration](EnvVarDBStatementTimeout), configutil.Lazy(&c.StatementTimeout)),
		configutil.Set(&c.SSLMode, configutil.Env[string](EnvVarDBSSLMode), configutil.Lazy(&c.SSLMode)),
		configutil.Set(&c.IdleConnections, configutil.Env[int](EnvVarDBIdleConnections), configutil.Lazy(&c.IdleConnections), configutil.Const(DefaultIdleConnections)),
		configutil.Set(&c.MaxConnections, configutil.Env[int](EnvVarDBMaxConnections), configutil.Lazy(&c.MaxConnections), configutil.Const(DefaultMaxConnections)),
		configutil.Set(&c.MaxLifetime, configutil.Env[time.Duration](EnvVarDBMaxLifetime), configutil.Lazy(&c.MaxLifetime), configutil.Const(DefaultMaxLifetime)),
		configutil.Set(&c.MaxIdleTime, configutil.Env[time.Duration](EnvVarDBMaxIdleTime), configutil.Lazy(&c.MaxIdleTime), configutil.Const(DefaultMaxIdleTime)),
		configutil.Set(&c.BufferPoolSize, configutil.Env[int](EnvVarDBBufferPoolSize), configutil.Lazy(&c.BufferPoolSize), configutil.Const(DefaultBufferPoolSize)),
		configutil.Set(&c.Dialect, configutil.Env[string](EnvVarDBDialect), configutil.Lazy(&c.Dialect), configutil.Const(string(DialectPostgres))),
	)
}

// CreateDSN creates a postgres connection string from the config.
func (c Config) CreateDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	host := c.Host
	if c.Port != "" {
		host = host + ":" + c.Port
	}

	dsn := &url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   c.Database,
	}

	if len(c.Username) > 0 {
		if len(c.Password) > 0 {
			dsn.User = url.UserPassword(c.Username, c.Password)
		} else {
			dsn.User = url.User(c.Username)
		}
	}

	queryArgs := url.Values{}
	if len(c.SSLMode) > 0 {
		queryArgs.Add("sslmode", c.SSLMode)
	}
	if c.ConnectTimeout > 0 {
		setTimeoutSeconds(queryArgs, "connect_timeout", c.ConnectTimeout)
	}
	if c.LockTimeout > 0 {
		setTimeoutMilliseconds(queryArgs, "lock_timeout", c.LockTimeout)
	}
	if c.StatementTimeout > 0 {
		setTimeoutMilliseconds(queryArgs, "statement_timeout", c.StatementTimeout)
	}
	if c.Schema != "" {
		queryArgs.Add("search_path", c.Schema)
	}
	if c.ApplicationName != "" {
		queryArgs.Add("application_name", c.ApplicationName)
	}

	dsn.RawQuery = queryArgs.Encode()
	return dsn.String()
}

// CreateLoggingDSN creates a postgres connection string from the config suitable for logging.
// It will not include the password.
func (c Config) CreateLoggingDSN() string {
	// NOTE: If we're provided a DSN, we _really_ want to have CreateDSN get
	//       skipped because it'll just dump the DSN raw.
	if c.DSN != "" {
		u, _ := url.Parse(c.DSN)
		u.User = url.User(u.User.Username())
		return u.String()
	}
	// NOTE: Since `c` is a value receiver, we can modify it without
	//       mutating the actual value.
	c.Password = ""
	return c.CreateDSN()
}

// Validate validates that user-provided values are valid, e.g. that timeouts
// can be exactly rounded into a multiple of a given base value.
func (c Config) Validate() error {
	if c.ConnectTimeout.Round(time.Second) != c.ConnectTimeout {
		return fmt.Errorf("invalid configuration value; connect_timeout=%s; %w", c.ConnectTimeout, ErrDurationConversion)
	}
	if c.LockTimeout.Round(time.Millisecond) != c.LockTimeout {
		return fmt.Errorf("invalid configuration value; lock_timeout=%s; %w", c.LockTimeout, ErrDurationConversion)
	}
	if c.StatementTimeout.Round(time.Millisecond) != c.StatementTimeout {
		return fmt.Errorf("invalid configuration value; statement_timeout=%s; %w", c.StatementTimeout, ErrDurationConversion)
	}

	return nil
}

// ValidateProduction validates production configuration for the config.
func (c Config) ValidateProduction() error {
	if !(len(c.SSLMode) == 0 ||
		strings.EqualFold(c.SSLMode, SSLModeRequire) ||
		strings.EqualFold(c.SSLMode, SSLModeVerifyCA) ||
		strings.EqualFold(c.SSLMode, SSLModeVerifyFull)) {
		return ErrUnsafeSSLMode
	}
	if len(c.Username) == 0 {
		return ErrUsernameUnset
	}
	if len(c.Password) == 0 {
		return ErrPasswordUnset
	}
	return c.Validate()
}

// setTimeoutMilliseconds sets a timeout value in connection string query parameters.
//
// Valid units for this parameter in PostgresSQL are "ms", "s", "min", "h"
// and "d" and the value should be between 0 and 2147483647ms. We explicitly
// cast to milliseconds but leave validation on the value to PostgreSQL.
//
// See:
// - https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-LOCK-TIMEOUT
// - https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
func setTimeoutMilliseconds(q url.Values, name string, d time.Duration) {
	ms := d.Round(time.Millisecond) / time.Millisecond
	q.Add(name, fmt.Sprintf("%dms", ms))
}

// setTimeoutSeconds sets a timeout value in connection string query parameters.
//
// This timeout is expected to be an exact number of seconds (as an integer)
// so we convert `d` to an integer first and set the value as a query parameter
// without units.
//
// See:
// - https://www.postgresql.org/docs/10/libpq-connect.html#LIBPQ-CONNECT-CONNECT-TIMEOUT
func setTimeoutSeconds(q url.Values, name string, d time.Duration) {
	s := d.Round(time.Second) / time.Second
	q.Add(name, fmt.Sprintf("%d", s))
}
