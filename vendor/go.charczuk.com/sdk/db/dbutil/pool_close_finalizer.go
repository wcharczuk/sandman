package dbutil

import (
	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/errutil"
)

// PoolCloseFinalizer is intended to be used in `defer` blocks with a named
// `error` return. It ensures a pool is closed after usage in contexts where
// a "limited use" pool is created.
//
// > func queries() (err error) {
// > 	var pool *db.Connection
// > 	defer func() {
// > 		err = db.PoolCloseFinalizer(pool, err)
// > 	}()
// > 	// ...
// > }
func PoolCloseFinalizer(pool *db.Connection, err error) error {
	if pool == nil {
		return err
	}
	closeErr := pool.Close()
	if closeErr != nil {
		return errutil.Append(err, closeErr)
	}
	return err
}
