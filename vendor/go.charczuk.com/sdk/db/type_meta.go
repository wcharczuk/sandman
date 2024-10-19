package db

import (
	"reflect"
	"strings"
)

// TypeMetaFor returns the TypeMeta for an object.
func TypeMetaFor(object any) *TypeMeta {
	t := reflect.TypeOf(object)
	return NewTypeMetaFromColumns(generateColumnsForType(nil, t)...)
}

// NewTypeMetaFromColumns returns a new TypeMeta instance from a given list of columns.
func NewTypeMetaFromColumns(columns ...Column) *TypeMeta {
	cc := TypeMeta{
		columns: columns,
	}
	lookup := make(map[string]*Column)
	for i := 0; i < len(columns); i++ {
		col := &columns[i]
		lookup[col.ColumnName] = col
	}
	cc.lookup = lookup
	return &cc
}

// NewTypeMetaFromColumnsWithPrefix makes a new TypeMeta instance from a given list
// of columns with a column prefix.
func NewTypeMetaFromColumnsWithPrefix(columnPrefix string, columns ...Column) *TypeMeta {
	cc := TypeMeta{
		columns: columns,
	}
	lookup := make(map[string]*Column)
	for i := 0; i < len(columns); i++ {
		col := &columns[i]
		lookup[col.ColumnName] = col
	}
	cc.lookup = lookup
	cc.columnPrefix = columnPrefix
	return &cc
}

// TypeMeta represents the column metadata for a given struct.
type TypeMeta struct {
	columns      []Column
	lookup       map[string]*Column
	columnPrefix string

	autos          []*Column
	notAutos       []*Column
	readOnly       []*Column
	notReadOnly    []*Column
	primaryKeys    []*Column
	notPrimaryKeys []*Column
	uniqueKeys     []*Column
	notUniqueKeys  []*Column
	insertColumns  []*Column
	updateColumns  []*Column
}

// Len returns the number of columns.
func (cc *TypeMeta) Len() int {
	if cc == nil {
		return 0
	}
	return len(cc.columns)
}

// Add adds a column.
func (cc *TypeMeta) Add(c Column) {
	cc.columns = append(cc.columns, c)
	cc.lookup[c.ColumnName] = &c
}

// Remove removes a column (by column name) from the collection.
func (cc *TypeMeta) Remove(columnName string) {
	var newColumns []Column
	for _, c := range cc.columns {
		if c.ColumnName != columnName {
			newColumns = append(newColumns, c)
		}
	}
	cc.columns = newColumns
	delete(cc.lookup, columnName)
}

// Column returns a column by name is present in the collection.
func (cc *TypeMeta) Column(columnName string) (c *Column) {
	c = cc.lookup[columnName]
	return
}

// HasColumn returns if a column name is present in the collection.
func (cc *TypeMeta) HasColumn(columnName string) bool {
	_, hasColumn := cc.lookup[columnName]
	return hasColumn
}

// Copy creates a new TypeMeta instance and carries over an existing column prefix.
func (cc *TypeMeta) Copy() *TypeMeta {
	return NewTypeMetaFromColumnsWithPrefix(cc.columnPrefix, cc.columns...)
}

// CopyWithColumnPrefix applies a column prefix to column names and returns a new column collection.
func (cc *TypeMeta) CopyWithColumnPrefix(prefix string) *TypeMeta {
	return NewTypeMetaFromColumnsWithPrefix(prefix, cc.columns...)
}

// InsertColumns are non-auto, non-readonly columns.
func (cc *TypeMeta) InsertColumns() []*Column {
	if cc.insertColumns != nil {
		return cc.insertColumns
	}
	for index, col := range cc.columns {
		if !col.IsReadOnly && !col.IsAuto {
			cc.insertColumns = append(cc.insertColumns, &cc.columns[index])
		}
	}
	return cc.insertColumns
}

// UpdateColumns are non-primary key, non-readonly columns.
func (cc *TypeMeta) UpdateColumns() []*Column {
	if cc.updateColumns != nil {
		return cc.updateColumns
	}
	for index, col := range cc.columns {
		if !col.IsReadOnly && !col.IsPrimaryKey {
			cc.updateColumns = append(cc.updateColumns, &cc.columns[index])
		}
	}
	return cc.updateColumns
}

// PrimaryKeys are columns we use as where predicates and can't update.
func (cc *TypeMeta) PrimaryKeys() []*Column {
	if cc.primaryKeys != nil {
		return cc.primaryKeys
	}
	for index, col := range cc.columns {
		if col.IsPrimaryKey {
			cc.primaryKeys = append(cc.primaryKeys, &cc.columns[index])
		}
	}
	return cc.primaryKeys
}

// NotPrimaryKeys are columns we can update.
func (cc *TypeMeta) NotPrimaryKeys() []*Column {
	if cc.notPrimaryKeys != nil {
		return cc.notPrimaryKeys
	}

	for index, col := range cc.columns {
		if !col.IsPrimaryKey {
			cc.notPrimaryKeys = append(cc.notPrimaryKeys, &cc.columns[index])
		}
	}
	return cc.notPrimaryKeys
}

// UniqueKeys are columns we use as where predicates and can't update.
func (cc *TypeMeta) UniqueKeys() []*Column {
	if cc.uniqueKeys != nil {
		return cc.uniqueKeys
	}
	for index, col := range cc.columns {
		if col.IsUniqueKey {
			cc.uniqueKeys = append(cc.uniqueKeys, &cc.columns[index])
		}
	}
	return cc.uniqueKeys
}

// NotUniqueKeys are columns we can update.
func (cc *TypeMeta) NotUniqueKeys() []*Column {
	if cc.notUniqueKeys != nil {
		return cc.notUniqueKeys
	}

	for index, col := range cc.columns {
		if !col.IsUniqueKey {
			cc.notUniqueKeys = append(cc.notUniqueKeys, &cc.columns[index])
		}
	}
	return cc.notUniqueKeys
}

// Autos are columns we have to return the id of.
func (cc *TypeMeta) Autos() []*Column {
	if cc.autos != nil {
		return cc.autos
	}

	for index, col := range cc.columns {
		if col.IsAuto {
			cc.autos = append(cc.autos, &cc.columns[index])
		}
	}
	return cc.autos
}

// NotAutos are columns we don't have to return the id of.
func (cc *TypeMeta) NotAutos() []*Column {
	if cc.notAutos != nil {
		return cc.notAutos
	}

	for index, col := range cc.columns {
		if !col.IsAuto {
			cc.notAutos = append(cc.notAutos, &cc.columns[index])
		}
	}
	return cc.notAutos
}

// ReadOnly are columns that we don't have to insert upon Create().
func (cc *TypeMeta) ReadOnly() []*Column {
	if cc.readOnly != nil {
		return cc.readOnly
	}

	for index, col := range cc.columns {
		if col.IsReadOnly {
			cc.readOnly = append(cc.readOnly, &cc.columns[index])
		}
	}
	return cc.readOnly
}

// NotReadOnly are columns that we have to insert upon Create().
func (cc *TypeMeta) NotReadOnly() []*Column {
	if cc.notReadOnly != nil {
		return cc.notReadOnly
	}

	for index, col := range cc.columns {
		if !col.IsReadOnly {
			cc.notReadOnly = append(cc.notReadOnly, &cc.columns[index])
		}
	}
	return cc.notReadOnly
}

// Columns returns the colummns
func (cc *TypeMeta) Columns() (output []*Column) {
	output = make([]*Column, len(cc.columns))
	for index := range cc.columns {
		output[index] = &cc.columns[index]
	}
	return
}

// Lookup gets the column name lookup.
func (cc *TypeMeta) Lookup() map[string]*Column {
	if len(cc.columnPrefix) != 0 {
		lookup := map[string]*Column{}
		for key, value := range cc.lookup {
			lookup[cc.columnPrefix+key] = value
		}
		return lookup
	}
	return cc.lookup
}

//
// helpers
//

// ColumnsZero returns unset fields on an instance that correspond to fields in the column collection.
func ColumnsZero(cols []*Column, instance any) (output []*Column) {
	objValue := reflectValue(instance)
	var fieldValue reflect.Value
	for index, c := range cols {
		fieldValue = objValue.Field(c.Index)
		if fieldValue.IsZero() {
			output = append(output, cols[index])
		}
	}
	return
}

// ColumnsNotZero returns set fields on an instance that correspond to fields in the column collection.
func ColumnsNotZero(cols []*Column, instance any) (output []*Column) {
	objValue := reflectValue(instance)
	var fieldValue reflect.Value
	for index, c := range cols {
		fieldValue = objValue.Field(c.Index)
		if !fieldValue.IsZero() {
			output = append(output, cols[index])
		}
	}
	return
}

// ColumnNames returns the string names for all the columns in the collection.
func ColumnNames(cols []*Column) []string {
	names := make([]string, len(cols))
	for x := 0; x < len(cols); x++ {
		c := cols[x]
		names[x] = c.ColumnName
	}
	return names
}

// ColumnNamesCSV returns a csv of column names.
func ColumnNamesCSV(cols []*Column) string {
	return strings.Join(ColumnNames(cols), ", ")
}

// ColumnNamesWithPrefix returns the string names for all the columns in the
// collection with a given disambiguation prefix.
func ColumnNamesWithPrefix(cols []*Column, columnPrefix string) []string {
	names := make([]string, len(cols))
	for x := 0; x < len(cols); x++ {
		c := cols[x]
		if len(columnPrefix) != 0 {
			names[x] = columnPrefix + c.ColumnName
		} else {
			names[x] = c.ColumnName
		}
	}
	return names
}

// ColumnNamesWithPrefixCSV returns a csv of column names with a given disambiguation prefix.
func ColumnNamesWithPrefixCSV(cols []*Column, columnPrefix string) string {
	return strings.Join(ColumnNamesWithPrefix(cols, columnPrefix), ", ")
}

// ColumnNamesFromAlias returns the string names for all the columns in the collection.
func ColumnNamesFromAlias(cols []*Column, tableAlias string) []string {
	names := make([]string, len(cols))
	for x := 0; x < len(cols); x++ {
		c := cols[x]
		names[x] = tableAlias + "." + c.ColumnName
	}
	return names
}

// ColumnNamesCSV returns a csv of column names.
func ColumnNamesFromAliasCSV(cols []*Column, tableAlias string) string {
	return strings.Join(ColumnNamesFromAlias(cols, tableAlias), ", ")
}

// ColumnNamesWithPrefixFromAlias returns the string names for all the columns in the collection.
func ColumnNamesWithPrefixFromAlias(cols []*Column, columnPrefix, tableAlias string) []string {
	names := make([]string, len(cols))
	for x := 0; x < len(cols); x++ {
		c := cols[x]
		if columnPrefix != "" {
			names[x] = tableAlias + "." + c.ColumnName + " as " + columnPrefix + c.ColumnName
		} else {
			names[x] = tableAlias + "." + c.ColumnName
		}
	}
	return names
}

// ColumnNamesCSV returns a csv of column names.
func ColumnNamesWithPrefixFromAliasCSV(cols []*Column, columnPrefix, tableAlias string) string {
	return strings.Join(ColumnNamesWithPrefixFromAlias(cols, columnPrefix, tableAlias), ", ")
}

// ColumnValues returns the reflected value for all the columns on a given instance.
func ColumnValues(cols []*Column, instance any) (output []any) {
	value := reflectValue(instance)
	output = make([]any, len(cols))
	for x := 0; x < len(cols); x++ {
		c := cols[x]
		valueField := value.FieldByName(c.FieldName)
		if c.IsJSON {
			output[x] = JSON(valueField.Interface())
		} else {
			output[x] = valueField.Interface()
		}
	}
	return
}

// HasColumn returns if a column with a given name exists in the list.
func HasColumn(cols []*Column, name string) bool {
	for _, c := range cols {
		if c.ColumnName == name {
			return true
		}
	}
	return false
}

// filter a slice with a function.
func filter[T any](values []T, fn func(T) bool) (out []T) {
	out = make([]T, 0, len(values))
	for _, v := range values {
		if fn(v) {
			out = append(out, v)
		}
	}
	return
}

// newColumnCacheKey creates a cache key for a type.
func newColumnCacheKey(objectType reflect.Type) string {
	typeName := objectType.String()
	instance := reflect.New(objectType).Interface()
	if typed, ok := instance.(ColumnMetaCacheKeyProvider); ok {
		return typeName + "_" + typed.ColumnMetaCacheKey()
	}
	if typed, ok := instance.(TableNameProvider); ok {
		return typeName + "_" + typed.TableName()
	}
	return typeName
}

// generateColumnsForType generates a column list for a given type.
func generateColumnsForType(parent *Column, t reflect.Type) []Column {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var tableName string
	if parent != nil {
		tableName = parent.TableName
	} else {
		tableName = TableNameByType(t)
	}

	numFields := t.NumField()

	var cols []Column
	for index := 0; index < numFields; index++ {
		field := t.Field(index)
		col := NewColumnFromFieldTag(field)
		if col != nil {
			col.Parent = parent
			col.Index = index
			col.TableName = tableName
			if col.Inline && field.Anonymous { // if it's not anonymous, whatchu doin
				cols = append(cols, generateColumnsForType(col, col.FieldType)...)
			} else if !field.Anonymous {
				cols = append(cols, *col)
			}
		}
	}

	return cols
}
