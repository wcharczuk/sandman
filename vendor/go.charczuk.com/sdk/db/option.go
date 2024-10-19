package db

import (
	"context"
	"database/sql"

	"go.charczuk.com/sdk/log"
)

// Option is an option for database connections.
type Option func(c *Connection) error

// OptConnection sets the underlying driver connection.
func OptConnection(conn *sql.DB) Option {
	return func(c *Connection) error {
		c.conn = conn
		return nil
	}
}

// OptLog enables logging on the connection.
func OptLog(log *log.Logger) Option {
	return func(c *Connection) error {
		c.onQuery = append(c.onQuery, func(qe QueryEvent) {
			log.WithGroup("DB").Info("query", qe)
			if qe.Err != nil {
				log.WithGroup("DB").Error("query_error", qe.Err)
			}
		})
		return nil
	}
}

// OptConfig sets the config on a connection.
func OptConfig(cfg Config) Option {
	return func(c *Connection) error {
		c.Config = cfg
		return nil
	}
}

// OptConfigFromEnv sets the config on a connection from the environment.
func OptConfigFromEnv() Option {
	return func(c *Connection) error {
		return (&c.Config).Resolve(context.Background())
	}
}

// OptEngine sets the connection engine.
// You must have this engine registered with database/sql.
func OptEngine(engine string) Option {
	return func(c *Connection) error {
		c.Config.Engine = engine
		return nil
	}
}

// OptHost sets the connection host.
func OptHost(host string) Option {
	return func(c *Connection) error {
		c.Config.Host = host
		return nil
	}
}

// OptPort sets the connection port.
func OptPort(port string) Option {
	return func(c *Connection) error {
		c.Config.Port = port
		return nil
	}
}

// OptDatabase sets the connection database.
func OptDatabase(database string) Option {
	return func(c *Connection) error {
		c.Config.Database = database
		return nil
	}
}

// OptUsername sets the connection ssl mode.
func OptUsername(username string) Option {
	return func(c *Connection) error {
		c.Config.Username = username
		return nil
	}
}

// OptPassword sets the connection ssl mode.
func OptPassword(password string) Option {
	return func(c *Connection) error {
		c.Config.Password = password
		return nil
	}
}

// OptSchema sets the connection schema path.
func OptSchema(schema string) Option {
	return func(c *Connection) error {
		c.Config.Schema = schema
		return nil
	}
}

// OptSSLMode sets the connection ssl mode.
func OptSSLMode(mode string) Option {
	return func(c *Connection) error {
		c.Config.SSLMode = mode
		return nil
	}
}

// OptDialect sets the connection dialect.
func OptDialect(dialect Dialect) Option {
	return func(c *Connection) error {
		c.Config.Dialect = string(dialect)
		return nil
	}
}
