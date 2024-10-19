package db

import (
	"context"
	"database/sql"
	"reflect"
	"sync"
	"time"
)

// New returns a new Connection.
// It will use very bare bones defaults for the config.
func New(options ...Option) (*Connection, error) {
	var c Connection
	var err error
	for _, opt := range options {
		if err = opt(&c); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// MustNew returns a new connection and panics on error.
func MustNew(options ...Option) *Connection {
	c, err := New(options...)
	if err != nil {
		panic(err)
	}
	return c
}

// Connection is a
type Connection struct {
	Config

	conn    *sql.DB
	bp      *BufferPool
	mcmu    sync.RWMutex
	mc      map[string]*TypeMeta
	onQuery []func(QueryEvent)
}

// RegisterOnQuery adds an on query listener.
func (dbc *Connection) RegisterOnQuery(fn func(QueryEvent)) {
	dbc.onQuery = append(dbc.onQuery, fn)
}

// Stats returns the stats for the connection.
func (dbc *Connection) Stats() sql.DBStats {
	return dbc.conn.Stats()
}

// Open returns a connection object, either a cached connection object or creating a new one in the process.
func (dbc *Connection) Open() error {
	// bail if we've already opened the connection.
	if dbc.conn != nil {
		return ErrConnectionAlreadyOpen
	}
	if dbc.Config.IsZero() {
		return ErrConfigUnset
	}
	if dbc.bp == nil {
		dbc.bp = NewBufferPool(dbc.Config.BufferPoolSize)
	}
	if dbc.mc == nil {
		dbc.mc = make(map[string]*TypeMeta)
	}

	dbConn, err := sql.Open(dbc.Config.Engine, dbc.Config.CreateDSN())
	if err != nil {
		return err
	}
	dbc.conn = dbConn
	dbc.conn.SetConnMaxLifetime(dbc.Config.MaxLifetime)
	dbc.conn.SetConnMaxIdleTime(dbc.Config.MaxIdleTime)
	dbc.conn.SetMaxIdleConns(dbc.Config.IdleConnections)
	dbc.conn.SetMaxOpenConns(dbc.Config.MaxConnections)
	return nil
}

// Close implements a closer.
func (dbc *Connection) Close() error {
	return dbc.conn.Close()
}

// TypeMeta returns the TypeMeta for an object.
func (dbc *Connection) TypeMeta(object any) *TypeMeta {
	objectType := reflect.TypeOf(object)
	return dbc.TypeMetaFromType(newColumnCacheKey(objectType), objectType)
}

// TypeMetaFromType reflects a reflect.Type into a column collection.
// The results of this are cached for speed.
func (dbc *Connection) TypeMetaFromType(identifier string, t reflect.Type) *TypeMeta {
	dbc.mcmu.RLock()
	if value, ok := dbc.mc[identifier]; ok {
		dbc.mcmu.RUnlock()
		return value
	}
	dbc.mcmu.RUnlock()

	dbc.mcmu.Lock()
	defer dbc.mcmu.Unlock()

	// check again once exclusive is acquired
	// because we may have been waiting on another routine
	// to set the identifier.
	if value, ok := dbc.mc[identifier]; ok {
		return value
	}

	metadata := NewTypeMetaFromColumns(generateColumnsForType(nil, t)...)
	dbc.mc[identifier] = metadata
	return metadata
}

// BeginTx starts a new transaction in a givent context.
func (dbc *Connection) BeginTx(ctx context.Context, opts ...func(*sql.TxOptions)) (*sql.Tx, error) {
	if dbc.conn == nil {
		return nil, ErrConnectionClosed
	}
	var txOptions sql.TxOptions
	for _, opt := range opts {
		opt(&txOptions)
	}
	tx, err := dbc.conn.BeginTx(ctx, &txOptions)
	return tx, err
}

// Prepare prepares a statement within a given context.
// If a tx is provided, the tx is the target for the prepare.
// This will trigger tracing on prepare.
func (dbc *Connection) Prepare(ctx context.Context, statement string) (stmt *sql.Stmt, err error) {
	stmt, err = dbc.conn.PrepareContext(ctx, statement)
	return
}

// Invoke returns a new invocation.
func (dbc *Connection) Invoke(options ...InvocationOption) *Invocation {
	i := Invocation{
		conn:    dbc,
		ctx:     context.Background(),
		started: time.Now().UTC(),
	}
	// we do a separate nil check here
	// so that we don't end up with a non-nil pointer to nil
	if dbc.conn != nil {
		i.db = dbc.conn
	}
	for _, option := range options {
		option(&i)
	}
	return &i
}

// Exec is a helper stub for .Invoke(...).Exec(...).
func (dbc *Connection) Exec(statement string, args ...interface{}) (sql.Result, error) {
	return dbc.Invoke().Exec(statement, args...)
}

// ExecContext is a helper stub for .Invoke(OptContext(ctx)).Exec(...).
func (dbc *Connection) ExecContext(ctx context.Context, statement string, args ...interface{}) (sql.Result, error) {
	return dbc.Invoke(OptContext(ctx)).Exec(statement, args...)
}

// Query is a helper stub for .Invoke(...).Query(...).
func (dbc *Connection) Query(statement string, args ...interface{}) *Query {
	return dbc.Invoke().Query(statement, args...)
}

// QueryContext is a helper stub for .Invoke(OptContext(ctx)).Query(...).
func (dbc *Connection) QueryContext(ctx context.Context, statement string, args ...interface{}) *Query {
	return dbc.Invoke(OptContext(ctx)).Query(statement, args...)
}
