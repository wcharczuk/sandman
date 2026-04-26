package db

import "errors"

var (
	// ErrDestinationNotStruct is an exception class.
	ErrDestinationNotStruct = errors.New("destination object is not a struct")
	// ErrConfigUnset is an exception class.
	ErrConfigUnset = errors.New("config is unset")
	// ErrUnsafeSSLMode is an error indicating unsafe ssl mode in production.
	ErrUnsafeSSLMode = errors.New("unsafe ssl mode in prodlike environment")
	// ErrUsernameUnset is an error indicating there is no username set in a prodlike environment.
	ErrUsernameUnset = errors.New("username is unset in prodlike environment")
	// ErrPasswordUnset is an error indicating there is no password set in a prodlike environment.
	ErrPasswordUnset = errors.New("password is unset in prodlike environment")
	// ErrDurationConversion is the error returned when a duration cannot be
	// converted to multiple of some base (e.g. milliseconds or seconds)
	// without round off.
	ErrDurationConversion = errors.New("cannot convert duration")
	// ErrConnectionAlreadyOpen is an error indicating the db connection was already opened.
	ErrConnectionAlreadyOpen = errors.New("the connection is already opened")
	// ErrConnectionClosed is an error indicating the db connection hasn't been opened.
	ErrConnectionClosed = errors.New("the connection is closed, or is being used before opened")
	// ErrContextUnset is an error indicating the context on the invocation isn't set.
	ErrContextUnset = errors.New("the context is unset; cannot continue")
	// ErrCollectionNotSlice is an error returned by OutMany if the destination is not a slice.
	ErrCollectionNotSlice = errors.New("outmany destination collection is not a slice")
	// ErrInvalidIDs is an error returned by Get if the ids aren't provided.
	ErrInvalidIDs = errors.New("invalid `ids` parameter")
	// ErrNoPrimaryKey is an error returned by a number of operations that depend on a primary key.
	ErrNoPrimaryKey = errors.New("no primary key on object")
	// ErrRowsNotColumnsProvider is returned by `PopulateByName` if you do not pass in `sql.Rows` as the scanner.
	ErrRowsNotColumnsProvider = errors.New("rows is not a columns provider")
	// ErrTooManyRows is returned by Out if there is more than one row returned by the query
	ErrTooManyRows = errors.New("too many rows returned to map to single object")
	// ErrNetwork is a grouped error for network issues.
	ErrNetwork = errors.New("network error")
)
