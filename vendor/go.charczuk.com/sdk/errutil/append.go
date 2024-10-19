package errutil

// Append appends errors together, creating a multi-error.
//
// Errors that are already multi-errors will be appended
// as is, creating a jagged error list.
//
// To have a one-dimensional resulting error list, use `AppendFlat(...)`.
func Append(err error, errs ...error) error {
	if err == nil && len(errs) == 0 {
		return nil
	}
	var output Multi
	switch typedErr := err.(type) {
	case Multi:
		output = typedErr
	default:
		if typedErr != nil {
			output = Multi{typedErr}
		}
	}
	for _, e := range errs {
		if e == nil {
			continue
		}
		output = append(output, e)
	}
	if len(output) > 0 {
		return output
	}
	return nil
}

// AppendFlat appends errors together, creating a multi-error.
//
// It differs from `Append(...)` in that it will flatten errors in the
// list that are already `Multi` errors, creating a one dimensional
// return list instead of a jagged return list.
func AppendFlat(err error, errs ...error) error {
	if err == nil && len(errs) == 0 {
		return nil
	}
	var output Multi
	switch typedErr := err.(type) {
	case Multi:
		output = typedErr
	default:
		if typedErr != nil {
			output = Multi{typedErr}
		}
	}
	for _, e := range errs {
		if e == nil {
			continue
		}
		output = append(output, flatten(e)...)
	}

	if len(output) > 0 {
		return output
	}
	return nil
}

// flatten takes _any_ error of any mishapen form
// and returns a single level error for it.
func flatten(err error) Multi {
	var output Multi
	switch typedErr := err.(type) {
	case Multi:
		for _, inner := range typedErr {
			output = append(output, flatten(inner)...)
		}
	default:
		if typedErr != nil {
			output = Multi{typedErr}
		}
	}
	return output
}
