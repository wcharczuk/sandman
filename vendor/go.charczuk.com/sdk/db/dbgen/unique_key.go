package dbgen

import (
	"strings"

	"go.charczuk.com/sdk/db"
)

// UniqueKey generates an alter table statement that adds a unique key constriant to a table.
func UniqueKey(obj any, columns ...string) string {
	for _, column := range columns {
		mustHaveColumn(obj, column)
	}

	objTableName := db.TableName(obj)
	columnLabels := strings.Join(columns, "_")
	columnCSV := strings.Join(columns, ",")
	return f(`ALTER TABLE {{ .objTableName }} ADD CONSTRAINT uk_{{ .objTableName }}_{{ .columnLabels }} UNIQUE ({{ .columnCSV }})`, v{
		"objTableName": objTableName,
		"columnLabels": columnLabels,
		"columnCSV":    columnCSV,
	})
}
