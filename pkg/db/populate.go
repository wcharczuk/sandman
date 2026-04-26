package db

import (
	"database/sql"
	"fmt"
	"reflect"
)

// PopulateByName sets the values of an object from the values of a sql.Rows object using column names.
func PopulateByName(object any, row Rows, cols *TypeMeta) error {
	rowColumns, err := row.Columns()
	if err != nil {
		return err
	}

	var values = make([]interface{}, len(rowColumns))
	var columnLookup = cols.Lookup()
	for i, name := range rowColumns {
		if col, ok := columnLookup[name]; ok {
			initColumnValue(i, values, col)
		} else {
			var value any
			values[i] = &value
		}
	}

	err = row.Scan(values...)
	if err != nil {
		return err
	}

	var colName string
	var field *Column
	var ok bool

	objectValue := reflectValue(object)
	for i, v := range values {
		colName = rowColumns[i]
		if field, ok = columnLookup[colName]; ok {
			err = field.SetValueReflected(objectValue, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsPopulatable returns if an object is populatable
func IsPopulatable(object any) (isPopulatable bool) {
	_, isPopulatable = object.(Populatable)
	return
}

// IsScanner returns if an scalar value is scannable
func IsScanner(object any) (isScanner bool) {
	_, isScanner = object.(sql.Scanner)
	return
}

// PopulateInOrder sets the values of an object in order from a sql.Rows object.
// Only use this method if you're certain of the column order. It is faster than populateByName.
// Optionally if your object implements Populatable this process will be skipped completely, which is even faster.
func PopulateInOrder(object any, row Scanner, cols []*Column) (err error) {
	var values = make([]interface{}, len(cols))

	for i, col := range cols {
		initColumnValue(i, values, col)
	}
	if err = row.Scan(values...); err != nil {
		return err
	}

	objectValue := reflectValue(object)
	var field *Column
	for i, v := range values {
		field = cols[i]
		if err = field.SetValueReflected(objectValue, v); err != nil {
			return
		}
	}

	return
}

// Zero resets an object.
func Zero(object any) error {
	objectValue := reflect.ValueOf(object)
	if !objectValue.Elem().CanSet() {
		return fmt.Errorf("zero; cannot set object, did you pass a reference?")
	}
	objectValue.Elem().Set(reflect.Zero(objectValue.Type().Elem()))
	return nil
}

// initColumnValue inserts the correct placeholder in the scan array of values.
// it will use `sql.Null` forms where appropriate.
// JSON fields are implicitly nullable.
func initColumnValue(index int, values []interface{}, col *Column) {
	if col.IsJSON {
		values[index] = &sql.NullString{}
	} else if col.FieldType.Kind() == reflect.Ptr {
		values[index] = reflect.New(col.FieldType).Interface()
	} else {
		values[index] = reflect.New(reflect.PtrTo(col.FieldType)).Interface()
	}
}
