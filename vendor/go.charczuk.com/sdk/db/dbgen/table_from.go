package dbgen

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"go.charczuk.com/sdk/db"
	"go.charczuk.com/sdk/db/migration"
)

// TableFrom returns a migration step for a given type to initialize
// a database with a table for that type.
//
// Note it does _not_ cover extra stuff like indices and constraints
// and other table features; for those you'll want to generate those with
// dedicated helpers and pass them in the `extra ...string` varaidic argument.
func TableFrom(obj any, extra ...string) *migration.Step {
	tableName := db.TableName(obj)

	var columns []string
	var primaryKeys []string
	var columnDefinition string
	var leadingComma string
	for index, column := range db.TypeMetaFor(obj).Columns() {
		if index > 0 {
			leadingComma = ", "
		}

		if column.IsPrimaryKey {
			primaryKeys = append(primaryKeys, column.ColumnName)
			if column.IsAuto {
				columnDefinition = fmt.Sprintf("%s%s %s NOT NULL DEFAULT %s", leadingComma, column.ColumnName, dbTypeForFieldType(column.FieldType), dbAutoForFieldType(column.FieldType))
			} else {
				columnDefinition = fmt.Sprintf("%s%s %s NOT NULL", leadingComma, column.ColumnName, dbTypeForFieldType(column.FieldType))
			}
		} else if column.IsAuto {
			columnDefinition = fmt.Sprintf("%s%s %s NOT NULL DEFAULT %s", leadingComma, column.ColumnName, dbTypeForFieldType(column.FieldType), dbAutoForFieldType(column.FieldType))
		} else if column.IsJSON {
			columnDefinition = fmt.Sprintf("%s%s %s", leadingComma, column.ColumnName, "JSONB")
		} else {
			if column.FieldType.Kind() == reflect.Ptr {
				columnDefinition = fmt.Sprintf("%s%s %s", leadingComma, column.ColumnName, dbTypeForFieldType(column.FieldType))
			} else {
				columnDefinition = fmt.Sprintf("%s%s %s NOT NULL", leadingComma, column.ColumnName, dbTypeForFieldType(column.FieldType))
			}
		}
		columns = append(columns, columnDefinition)
	}

	var statements []string
	statements = append(statements, f(`CREATE TABLE {{ .TableName }} (
{{- range $index, $column := .Columns }}
	{{ $column }}
{{- end }}
)`, v{"TableName": tableName, "Columns": columns}),
	)

	if len(primaryKeys) > 0 {
		statements = append(statements, f(`ALTER TABLE {{ .TableName }} ADD CONSTRAINT {{ .ConstraintName }} PRIMARY KEY ({{.Columns}})`, v{
			"TableName":      tableName,
			"ConstraintName": fmt.Sprintf("pk_%s_%s", tableName, strings.Join(primaryKeys, "_")),
			"Columns":        strings.Join(primaryKeys, ","),
		}))
	}
	return migration.NewStep(
		migration.TableNotExists(tableName),
		migration.Statements(
			append(statements, extra...)...,
		),
	)
}

func dbTypeForFieldType(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.String:
		return "TEXT"
	case reflect.Int32:
		return "INT"
	case reflect.Int, reflect.Int64:
		return "BIGINT"
	case reflect.Uint32:
		return "INT"
	case reflect.Uint, reflect.Uint64:
		return "BIGINT"
	case reflect.Float32:
		return "REAL"
	case reflect.Float64:
		return "DOUBLE PRECISION"
	default:
	}
	switch t.Name() {
	case "Time":
		if t.PkgPath() == "time" {
			return "TIMESTAMP"
		}
	case "Duration":
		if t.PkgPath() == "time" {
			return "INTERVAL"
		}
	case "UUID":
		if t.PkgPath() == "go.charczuk.com/sdk/uuid" {
			return "UUID"
		}
	}
	panic(fmt.Sprintf("unknown field type for db: %s/%s", t.PkgPath(), t.Name()))
}

func dbAutoForFieldType(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Name() {
	case "Time":
		if t.PkgPath() == "time" {
			return "current_timestamp"
		}
	case "UUID":
		if t.PkgPath() == "go.charczuk.com/sdk/uuid" {
			return "gen_random_uuid()"
		}
	default:
	}
	panic(fmt.Sprintf("unknown auto field type for db: %s/%s", t.PkgPath(), t.Name()))
}

// v is a map between string and anything.
//
// It is typically used as an argument to `f(...)` and
// lets you associate values with keys.
type v map[string]any

// f uses text/template to format a given string
func f(format string, args any) string {
	t, err := template.New("").Parse(format)
	if err != nil {
		return ""
	}

	buf := new(bytes.Buffer)
	_ = t.Execute(buf, args)
	return buf.String()
}
