package db

import "database/sql"

// ExecAffectedAny is a helper that returns if a given
// Exec result affected any rows.
func ExecAffectedAny(res sql.Result, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// ExecErr is a helper that just returns the exec err.
func ExecErr(_ sql.Result, err error) error {
	return err
}
