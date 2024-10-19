package dbgen

import (
	"fmt"

	"go.charczuk.com/sdk/db"
)

func mustHaveColumn(obj any, col string) {
	if !hasColumn(obj, col) {
		panic(fmt.Sprintf("object %T does not have a column with name %q", obj, col))
	}
}

func hasColumn(obj any, col string) bool {
	return db.TypeMetaFor(obj).HasColumn(col)
}
