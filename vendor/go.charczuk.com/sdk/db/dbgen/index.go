package dbgen

import (
	"strings"

	"go.charczuk.com/sdk/db"
)

// Index generates an create index statement.
func Index(obj any, columns ...string) string {
	for _, column := range columns {
		mustHaveColumn(obj, column)
	}

	objTableName := db.TableName(obj)
	columnLabels := strings.Join(columns, "_")
	columnCSV := strings.Join(columns, ",")
	return f(`CREATE INDEX ix_{{ .objTableName }}_{{ .columnLabels }} ON {{ .objTableName }}({{ .columnCSV }})`, v{
		"objTableName": objTableName,
		"columnLabels": columnLabels,
		"columnCSV":    columnCSV,
	})
}
