package dbgen

import (
	"go.charczuk.com/sdk/db"
)

// ForeignKey generates an alter table statement that adds a foreign key to a table.
func ForeignKey(obj any, objCol string, reference any, referenceCol string) string {
	mustHaveColumn(obj, objCol)
	mustHaveColumn(reference, referenceCol)
	objTableName := db.TableName(obj)
	referenceTableName := db.TableName(reference)
	return f(`ALTER TABLE {{ .objTableName }} ADD CONSTRAINT fk_{{ .objTableName }}_{{ .objCol }} FOREIGN KEY ({{ .objCol }}) REFERENCES {{ .referenceTableName }}({{ .referenceCol }})`, v{
		"objTableName":       objTableName,
		"objCol":             objCol,
		"referenceTableName": referenceTableName,
		"referenceCol":       referenceCol,
	})
}
