package db

import (
	"database/sql"
	"errors"
	"reflect"

	"go.charczuk.com/sdk/errutil"
)

// Query is the intermediate result of a query.
type Query struct {
	inv  *Invocation
	err  error
	stmt string
	args []any
}

// Do runs a given query, yielding the raw results.
func (q *Query) Do() (rows *sql.Rows, err error) {
	defer func() {
		err = q.finish(recover(), nil, err)
	}()
	rows, err = q.query()
	return
}

// Any returns if there are any results for the query.
func (q *Query) Any() (found bool, err error) {
	var rows *sql.Rows
	defer func() {
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()
	rows, err = q.query()
	if err != nil {
		return
	}
	found = rows.Next()
	return
}

// None returns if there are no results for the query.
func (q *Query) None() (notFound bool, err error) {
	var rows *sql.Rows
	defer func() {
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()
	rows, err = q.query()
	if err != nil {
		return
	}
	notFound = !rows.Next()
	return
}

// Scan writes the results to a given set of local variables.
// It returns if the query produced a row, and returns `ErrTooManyRows` if there
// are multiple row results.
func (q *Query) Scan(args ...any) (found bool, err error) {
	var rows *sql.Rows
	defer func() {
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()

	rows, err = q.query()
	if err != nil {
		return
	}
	found, err = Scan(rows, args...)
	return
}

// Out writes the query result to a single object via. reflection mapping. If there is more than one result, the first
// result is mapped to to object, and ErrTooManyRows is returned. Out() will apply column values for any colums
// in the row result to the object, potentially zeroing existing values out.
func (q *Query) Out(object any) (found bool, err error) {
	var rows *sql.Rows
	defer func() {
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()

	rows, err = q.query()
	if err != nil {
		return
	}
	sliceType := reflectType(object)
	if sliceType.Kind() != reflect.Struct {
		err = ErrDestinationNotStruct
		return
	}
	columnMeta := q.inv.conn.TypeMeta(object)
	if rows.Next() {
		found = true
		if populatable, ok := object.(Populatable); ok {
			err = populatable.Populate(rows)
		} else {
			err = PopulateByName(object, rows, columnMeta)
		}
		if err != nil {
			return
		}
	} else if err = Zero(object); err != nil {
		return
	}
	if rows.Next() {
		err = ErrTooManyRows
	}
	return
}

// OutMany writes the query results to a slice of objects.
func (q *Query) OutMany(collection any) (err error) {
	var rows *sql.Rows
	defer func() {
		// err = q.finish(nil, nil, err)
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()

	rows, err = q.query()
	if err != nil {
		return
	}

	sliceType := reflectType(collection)
	if sliceType.Kind() != reflect.Slice {
		err = ErrCollectionNotSlice
		return
	}

	sliceInnerType := reflectSliceType(collection)
	collectionValue := reflectValue(collection)
	v := makeNew(sliceInnerType)
	isStruct := sliceInnerType.Kind() == reflect.Struct
	var meta *TypeMeta
	if isStruct {
		meta = q.inv.conn.TypeMetaFromType(newColumnCacheKey(sliceInnerType), sliceInnerType)
	}

	isPopulatable := IsPopulatable(v)

	var didSetRows bool
	for rows.Next() {
		newObj := makeNew(sliceInnerType)
		if isPopulatable {
			err = newObj.(Populatable).Populate(rows)
		} else if isStruct {
			err = PopulateByName(newObj, rows, meta)
		} else {
			err = rows.Scan(newObj)
		}
		if err != nil {
			return
		}

		newObjValue := reflectValue(newObj)
		collectionValue.Set(reflect.Append(collectionValue, newObjValue))
		didSetRows = true
	}

	// this initializes the slice if we didn't add elements to it.
	if !didSetRows {
		collectionValue.Set(reflect.MakeSlice(sliceType, 0, 0))
	}
	return
}

// Each executes the consumer for each result of the query (one to many).
func (q *Query) Each(consumer RowsConsumer) (err error) {
	var rows *sql.Rows
	defer func() {
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()

	rows, err = q.query()
	if err != nil {
		return
	}

	err = Each(rows, consumer)
	return
}

// First executes the consumer for the first result of a query.
// It returns `ErrTooManyRows` if more than one result is returned.
func (q *Query) First(consumer RowsConsumer) (found bool, err error) {
	var rows *sql.Rows
	defer func() {
		err = q.finish(recover(), nil, err)
		err = q.rowsClose(rows, err)
	}()
	rows, err = q.query()
	if err != nil {
		return
	}
	found, err = First(rows, consumer)
	return
}

// --------------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------------

func (q *Query) rowsClose(rows *sql.Rows, err error) error {
	if rows == nil {
		return err
	}
	if closeErr := rows.Close(); closeErr != nil {
		return errutil.Append(err, closeErr)
	}
	return err
}

func (q *Query) query() (rows *sql.Rows, err error) {
	if q.err != nil {
		err = q.err
		return
	}

	var queryError error
	dbc := q.inv.db
	ctx := q.inv.ctx
	rows, queryError = dbc.QueryContext(ctx, q.stmt, q.args...)
	if queryError != nil && !errors.Is(queryError, sql.ErrNoRows) {
		err = queryError
	}
	return
}

func (q *Query) finish(r any, res sql.Result, err error) error {
	return q.inv.finish(q.stmt, r, res, err)
}

// Each iterates over a given result set, calling the rows consumer.
func Each(rows *sql.Rows, consumer RowsConsumer) (err error) {
	for rows.Next() {
		if err = consumer(rows); err != nil {
			return
		}
	}
	return
}

// First returns the first result of a result set to a consumer.
// If there are more than one row in the result, they are ignored.
func First(rows *sql.Rows, consumer RowsConsumer) (found bool, err error) {
	if found = rows.Next(); found {
		if err = consumer(rows); err != nil {
			return
		}
	}
	return
}

// Scan reads the first row from a resultset and scans it to a given set of args.
// If more than one row is returned it will return ErrTooManyRows.
func Scan(rows *sql.Rows, args ...any) (found bool, err error) {
	if rows.Next() {
		found = true
		if err = rows.Scan(args...); err != nil {
			return
		}
	}
	if rows.Next() {
		err = ErrTooManyRows
	}
	return
}
