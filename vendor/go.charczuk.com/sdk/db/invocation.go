package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.charczuk.com/sdk/errutil"
)

// InvocationOption is an option for invocations.
type InvocationOption func(*Invocation)

// OptLabel sets the Label on the invocation.
func OptLabel(label string) InvocationOption {
	return func(i *Invocation) {
		i.label = label
	}
}

// OptContext sets a context on an invocation.
func OptContext(ctx context.Context) InvocationOption {
	return func(i *Invocation) {
		i.ctx = ctx
	}
}

// OptCancel sets the context cancel func.
func OptCancel(cancel context.CancelFunc) InvocationOption {
	return func(i *Invocation) {
		i.cancel = cancel
	}
}

// OptTimeout sets a command timeout for the invocation.
func OptTimeout(d time.Duration) InvocationOption {
	return func(i *Invocation) {
		i.ctx, i.cancel = context.WithTimeout(i.ctx, d)
	}
}

// OptTx is an invocation option that sets the invocation transaction.
func OptTx(tx *sql.Tx) InvocationOption {
	return func(i *Invocation) {
		if tx != nil {
			i.db = tx
		}
	}
}

// OptInvocationDB is an invocation option that sets the underlying invocation db.
func OptInvocationDB(db DB) InvocationOption {
	return func(i *Invocation) {
		i.db = db
	}
}

// Invocation is a specific operation against a context.
type Invocation struct {
	conn    *Connection
	db      DB
	label   string
	ctx     context.Context
	cancel  func()
	started time.Time
}

// Exec executes a sql statement with a given set of arguments and returns the rows affected.
func (i *Invocation) Exec(statement string, args ...interface{}) (res sql.Result, err error) {
	statement, err = i.start(statement)
	if err != nil {
		return
	}
	defer func() { err = i.finish(statement, recover(), res, err) }()
	res, err = i.db.ExecContext(i.ctx, statement, args...)
	return
}

// Query returns a new query object for a given sql query and arguments.
func (i *Invocation) Query(statement string, args ...interface{}) *Query {
	q := &Query{
		inv:  i,
		args: args,
	}
	q.stmt, q.err = i.start(statement)
	return q
}

// Get returns a given object based on a group of primary key ids within a transaction.
func (i *Invocation) Get(object any, ids ...any) (found bool, err error) {
	if len(ids) == 0 {
		err = ErrInvalidIDs
		return
	}

	var queryBody, label string
	if label, queryBody, err = i.generateGet(object); err != nil {
		return
	}
	i.maybeSetLabel(label)
	return i.Query(queryBody, ids...).Out(object)
}

// GetMany returns objects matching a given array of keys.
//
// The order of the results will match the order of the keys.
func (i *Invocation) GetMany(collection any, ids ...any) (err error) {
	if len(ids) == 0 {
		err = ErrInvalidIDs
		return
	}
	var queryBody, label string
	if label, queryBody, err = i.generateGetMany(collection, len(ids)); err != nil {
		return
	}
	i.maybeSetLabel(label)
	return i.Query(queryBody, ids...).OutMany(collection)
}

// All returns all rows of an object mapped table wrapped in a transaction.
func (i *Invocation) All(collection interface{}) (err error) {
	label, queryBody := i.generateGetAll(collection)
	i.maybeSetLabel(label)
	return i.Query(queryBody).OutMany(collection)
}

// Create writes an object to the database within a transaction.
func (i *Invocation) Create(object any) (err error) {
	var queryBody, label string
	var insertCols, autos []*Column
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	label, queryBody, insertCols, autos = i.generateCreate(object)
	i.maybeSetLabel(label)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if len(autos) == 0 {
		if res, err = i.db.ExecContext(i.ctx, queryBody, ColumnValues(insertCols, object)...); err != nil {
			return
		}
		return
	}

	autoValues := i.autoValues(autos)
	if err = i.db.QueryRowContext(i.ctx, queryBody, ColumnValues(insertCols, object)...).Scan(autoValues...); err != nil {
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		return
	}
	return
}

// CreateIfNotExists writes an object to the database if it does not already exist within a transaction.
// This will _ignore_ auto columns, as they will always invalidate the assertion that there already exists
// a row with a given primary key set.
func (i *Invocation) CreateIfNotExists(object any) (err error) {
	var queryBody, label string
	var insertCols []*Column
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	label, queryBody, insertCols = i.generateCreateIfNotExists(object)
	i.maybeSetLabel(label)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	res, err = i.db.ExecContext(i.ctx, queryBody, ColumnValues(insertCols, object)...)
	return
}

// Update updates an object wrapped in a transaction. Returns whether or not any rows have been updated and potentially
// an error. If ErrTooManyRows is returned, it's important to note that due to https://github.com/golang/go/issues/7898,
// the Update HAS BEEN APPLIED. Its on the developer using UPDATE to ensure his tags are correct and/or execute it in a
// transaction and roll back on this error
func (i *Invocation) Update(object any) (updated bool, err error) {
	var queryBody, label string
	var pks, updateCols []*Column
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	label, queryBody, pks, updateCols = i.generateUpdate(object)
	i.maybeSetLabel(label)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	res, err = i.db.ExecContext(
		i.ctx,
		queryBody,
		append(ColumnValues(updateCols, object), ColumnValues(pks, object)...)...,
	)
	if err != nil {
		return
	}

	var rowCount int64
	rowCount, err = res.RowsAffected()
	if err != nil {
		return
	}
	if rowCount > 0 {
		updated = true
	}
	if rowCount > 1 {
		err = ErrTooManyRows
	}
	return
}

// Upsert inserts the object if it doesn't exist already (as defined by its primary keys) or updates it atomically.
// It returns `found` as true if the effect was an upsert, i.e. the pk was found.
func (i *Invocation) Upsert(object any) (err error) {
	var queryBody, label string
	var autos, upsertCols []*Column
	defer func() { err = i.finish(queryBody, recover(), nil, err) }()

	i.label, queryBody, autos, upsertCols = i.generateUpsert(object)
	i.maybeSetLabel(label)

	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	if len(autos) == 0 {
		if _, err = i.db.ExecContext(i.ctx, queryBody, ColumnValues(upsertCols, object)...); err != nil {
			return
		}
		return
	}
	autoValues := i.autoValues(autos)
	if err = i.db.QueryRowContext(i.ctx, queryBody, ColumnValues(upsertCols, object)...).Scan(autoValues...); err != nil {
		return
	}
	if err = i.setAutos(object, autos, autoValues); err != nil {
		return
	}
	return
}

// Exists returns a bool if a given object exists (utilizing the primary key columns if they exist) wrapped in a transaction.
func (i *Invocation) Exists(object any) (exists bool, err error) {
	var queryBody, label string
	var pks []*Column
	defer func() { err = i.finish(queryBody, recover(), nil, err) }()

	if label, queryBody, pks, err = i.generateExists(object); err != nil {
		return
	}
	i.maybeSetLabel(label)
	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	var value int
	if queryErr := i.db.QueryRowContext(i.ctx, queryBody, ColumnValues(pks, object)...).Scan(&value); queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		err = queryErr
		return
	}
	exists = value != 0
	return
}

// Delete deletes an object from the database wrapped in a transaction. Returns whether or not any rows have been deleted
// and potentially an error.
//
// If ErrTooManyRows is returned, it's important to note that due to https://github.com/golang/go/issues/7898
// the Delete HAS BEEN APPLIED on the current transaction. Its on the developer using Delete to ensure their
// tags are correct and/or ensure theit Tx rolls back on this error.
func (i *Invocation) Delete(object any) (deleted bool, err error) {
	var queryBody, label string
	var pks []*Column
	var res sql.Result
	defer func() { err = i.finish(queryBody, recover(), res, err) }()

	if label, queryBody, pks, err = i.generateDelete(object); err != nil {
		return
	}

	i.maybeSetLabel(label)
	queryBody, err = i.start(queryBody)
	if err != nil {
		return
	}
	res, err = i.db.ExecContext(i.ctx, queryBody, ColumnValues(pks, object)...)
	if err != nil {
		return
	}

	var rowCount int64
	rowCount, err = res.RowsAffected()
	if err != nil {
		return
	}
	if rowCount > 0 {
		deleted = true
	}
	if rowCount > 1 {
		err = ErrTooManyRows
	}
	return
}

func (i *Invocation) generateGet(object any) (cachePlan, queryBody string, err error) {
	tableName := TableName(object)

	cols := i.conn.TypeMeta(object)
	getCols := cols.NotReadOnly()
	pks := cols.PrimaryKeys()
	if len(pks) == 0 {
		err = ErrNoPrimaryKey
		return
	}

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range ColumnNamesWithPrefix(getCols, cols.columnPrefix) {
		queryBodyBuffer.WriteString(name)
		if i < (cols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")

	for i, pk := range pks {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (len(pks) - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	cachePlan = fmt.Sprintf("%s_get", tableName)
	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateGetMany(collection any, keys int) (
	cachePlan, queryBody string,
	err error,
) {
	collectionType := reflectSliceType(collection)
	tableName := TableNameByType(collectionType)

	cols := i.conn.TypeMetaFromType(tableName, reflectSliceType(collection))

	getCols := cols.NotReadOnly()
	pks := cols.PrimaryKeys()
	if len(pks) == 0 {
		err = ErrNoPrimaryKey
		return
	}
	pk := pks[0]

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range ColumnNamesWithPrefix(getCols, cols.columnPrefix) {
		queryBodyBuffer.WriteString(name)
		if i < (cols.Len() - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)

	queryBodyBuffer.WriteString(" WHERE ")
	queryBodyBuffer.WriteString(pk.ColumnName)
	queryBodyBuffer.WriteString(" IN (")

	for x := 0; x < keys; x++ {
		paramIndex := strconv.Itoa(x + 1)
		queryBodyBuffer.WriteString("$" + paramIndex)
		if x < (keys - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(")")
	cachePlan = fmt.Sprintf("%s_get_many", tableName)
	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateGetAll(collection interface{}) (statementLabel, queryBody string) {
	collectionType := reflectSliceType(collection)
	tableName := TableNameByType(collectionType)

	cols := i.conn.TypeMetaFromType(tableName, reflectSliceType(collection))

	// using `NotReadOnly` may seem confusing, but we don't want read only columns
	// because they are typically the result of a select clause
	// and not columns on the table represented by the type.
	getCols := cols.NotReadOnly()

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT ")
	for i, name := range ColumnNamesWithPrefix(getCols, cols.columnPrefix) {
		queryBodyBuffer.WriteString(name)
		if i < (len(getCols) - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(" FROM ")
	queryBodyBuffer.WriteString(tableName)

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_get_all"
	return
}

func (i *Invocation) generateCreate(object any) (statementLabel, queryBody string, insertCols, autos []*Column) {
	tableName := TableName(object)

	cols := i.conn.TypeMeta(object)
	insertCols = append(cols.InsertColumns(), ColumnsNotZero(cols.Autos(), object)...)
	autos = cols.Autos()

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range ColumnNamesWithPrefix(insertCols, cols.columnPrefix) {
		queryBodyBuffer.WriteString(name)
		if i < (len(insertCols) - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")
	for x := 0; x < len(insertCols); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (len(insertCols) - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(")")

	if len(autos) > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(ColumnNamesWithPrefixCSV(autos, cols.columnPrefix))
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_create"
	return
}

func (i *Invocation) generateCreateIfNotExists(object any) (statementLabel, queryBody string, insertCols []*Column) {
	cols := i.conn.TypeMeta(object)
	insertCols = append(cols.InsertColumns(), ColumnsNotZero(cols.Autos(), object)...)

	pks := cols.PrimaryKeys()
	tableName := TableName(object)

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")
	for i, name := range ColumnNamesWithPrefix(insertCols, cols.columnPrefix) {
		queryBodyBuffer.WriteString(name)
		if i < (len(insertCols) - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(") VALUES (")
	for x := 0; x < len(insertCols); x++ {
		queryBodyBuffer.WriteString("$" + strconv.Itoa(x+1))
		if x < (len(insertCols) - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}
	queryBodyBuffer.WriteString(")")

	if len(pks) > 0 {
		queryBodyBuffer.WriteString(" ON CONFLICT (")
		pkColumnNames := ColumnNamesWithPrefix(pks, cols.columnPrefix)
		for i, name := range pkColumnNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkColumnNames)-1 {
				queryBodyBuffer.WriteRune(',')
			}
		}
		queryBodyBuffer.WriteString(") DO NOTHING")
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_create_if_not_exists"
	return
}

func (i *Invocation) generateUpdate(object any) (statementLabel, queryBody string, pks, updateCols []*Column) {
	tableName := TableName(object)

	cols := i.conn.TypeMeta(object)

	pks = cols.PrimaryKeys()
	updateCols = cols.UpdateColumns()

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("UPDATE ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" SET ")

	var updateColIndex int
	var col *Column
	for ; updateColIndex < len(updateCols); updateColIndex++ {
		col = updateCols[updateColIndex]
		queryBodyBuffer.WriteString(col.ColumnName)
		queryBodyBuffer.WriteString(" = $" + strconv.Itoa(updateColIndex+1))
		if updateColIndex != (len(updateCols) - 1) {
			queryBodyBuffer.WriteRune(',')
		}
	}

	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+(updateColIndex+1)))

		if i < (len(pks) - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_update"
	return
}

func (i *Invocation) generateUpsert(object any) (statementLabel, queryBody string, autos, insertsWithAutos []*Column) {
	tableName := TableName(object)
	cols := i.conn.TypeMeta(object)

	inserts := cols.InsertColumns()
	updates := cols.UpdateColumns()
	insertsWithAutos = append(inserts, cols.Autos()...)

	pks := filter(insertsWithAutos, func(c *Column) bool { return c.IsPrimaryKey })
	notZero := ColumnsNotZero(cols.Columns(), object)

	// But we exclude auto primary keys that are not set. Auto primary keys that ARE set must be included in the insert
	// clause so that there is a collision. But keys that are not set must be excluded from insertsWithAutos so that
	// they are not passed as an extra parameter to ExecInContext later and are properly auto-generated
	for _, pk := range pks {
		if pk.IsAuto && !HasColumn(notZero, pk.ColumnName) {
			insertsWithAutos = filter(insertsWithAutos, func(c *Column) bool { return c.ColumnName == pk.ColumnName })
		}
	}

	tokenMap := map[string]string{}
	for i, col := range insertsWithAutos {
		tokenMap[col.ColumnName] = "$" + strconv.Itoa(i+1)
	}

	// autos are read out on insert (but only if unset)
	autos = ColumnsZero(cols.Autos(), object)
	pkNames := ColumnNames(pks)

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("INSERT INTO ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" (")

	skipComma := true
	for _, col := range insertsWithAutos {
		if !col.IsAuto || HasColumn(notZero, col.ColumnName) {
			if !skipComma {
				queryBodyBuffer.WriteRune(',')
			}
			skipComma = false
			queryBodyBuffer.WriteString(col.ColumnName)
		}
	}

	queryBodyBuffer.WriteString(") VALUES (")
	skipComma = true
	for _, col := range insertsWithAutos {
		if !col.IsAuto || HasColumn(notZero, col.ColumnName) {
			if !skipComma {
				queryBodyBuffer.WriteRune(',')
			}
			skipComma = false
			queryBodyBuffer.WriteString(tokenMap[col.ColumnName])
		}
	}

	queryBodyBuffer.WriteString(")")

	if len(pks) > 0 {
		queryBodyBuffer.WriteString(" ON CONFLICT (")

		for i, name := range pkNames {
			queryBodyBuffer.WriteString(name)
			if i < len(pkNames)-1 {
				queryBodyBuffer.WriteRune(',')
			}
		}
		if len(updates) > 0 {
			queryBodyBuffer.WriteString(") DO UPDATE SET ")

			for i, col := range updates {
				queryBodyBuffer.WriteString(col.ColumnName + " = " + tokenMap[col.ColumnName])
				if i < (len(updates) - 1) {
					queryBodyBuffer.WriteRune(',')
				}
			}
		} else {
			queryBodyBuffer.WriteString(") DO NOTHING ")
		}
	}
	if len(autos) > 0 {
		queryBodyBuffer.WriteString(" RETURNING ")
		queryBodyBuffer.WriteString(ColumnNamesCSV(autos))
	}

	queryBody = queryBodyBuffer.String()
	statementLabel = tableName + "_upsert"
	return
}

func (i *Invocation) generateExists(object any) (statementLabel, queryBody string, pks []*Column, err error) {
	tableName := TableName(object)
	pks = i.conn.TypeMeta(object).PrimaryKeys()
	if len(pks) == 0 {
		err = ErrNoPrimaryKey
		return
	}

	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("SELECT 1 FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (len(pks) - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	statementLabel = tableName + "_exists"
	queryBody = queryBodyBuffer.String()
	return
}

func (i *Invocation) generateDelete(object any) (statementLabel, queryBody string, pks []*Column, err error) {
	tableName := TableName(object)
	pks = i.conn.TypeMeta(object).PrimaryKeys()
	if len(pks) == 0 {
		err = ErrNoPrimaryKey
		return
	}
	queryBodyBuffer := i.conn.bp.Get()
	defer i.conn.bp.Put(queryBodyBuffer)

	queryBodyBuffer.WriteString("DELETE FROM ")
	queryBodyBuffer.WriteString(tableName)
	queryBodyBuffer.WriteString(" WHERE ")
	for i, pk := range pks {
		queryBodyBuffer.WriteString(pk.ColumnName)
		queryBodyBuffer.WriteString(" = ")
		queryBodyBuffer.WriteString("$" + strconv.Itoa(i+1))

		if i < (len(pks) - 1) {
			queryBodyBuffer.WriteString(" AND ")
		}
	}
	statementLabel = tableName + "_delete"
	queryBody = queryBodyBuffer.String()
	return
}

// --------------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------------

func (i *Invocation) maybeSetLabel(label string) {
	if i.label != "" {
		return
	}
	i.label = label
}

// autoValues returns references to the auto updatd fields for a given column collection.
func (i *Invocation) autoValues(autos []*Column) []interface{} {
	autoValues := make([]interface{}, len(autos))
	for i, autoCol := range autos {
		autoValues[i] = reflect.New(reflect.PtrTo(autoCol.FieldType)).Interface()
	}
	return autoValues
}

// setAutos sets the automatic values for a given object.
func (i *Invocation) setAutos(object any, autos []*Column, autoValues []any) (err error) {
	for index := 0; index < len(autoValues); index++ {
		err = autos[index].SetValue(object, autoValues[index])
		if err != nil {
			return
		}
	}
	return
}

// start runs on start steps.
func (i *Invocation) start(statement string) (string, error) {
	if i.db == nil {
		return "", ErrConnectionClosed
	}
	if i.ctx == nil {
		return "", ErrContextUnset
	}
	// there are a lot of steps here typically but we removed
	// most of the logger and statement interceptor stuff for simplicity.
	// we can add that back later.
	return statement, nil
}

// finish runs on complete steps.
func (i *Invocation) finish(statement string, r any, res sql.Result, err error) error {
	if i.cancel != nil {
		i.cancel()
	}
	if r != nil {
		err = errutil.Append(err, errutil.New(r))
	}
	if i.conn != nil && len(i.conn.onQuery) > 0 {
		qe := QueryEvent{
			Body:     statement,
			Elapsed:  time.Now().UTC().Sub(i.started),
			Username: i.conn.Config.Username,
			Database: i.conn.Config.Database,
			Engine:   i.conn.Config.Engine,
			Label:    i.label,
			Err:      errutil.New(err),
		}
		if res != nil {
			qe.RowsAffected, _ = res.RowsAffected()
		}
		for _, l := range i.conn.onQuery {
			l(qe)
		}
	}
	return err
}
