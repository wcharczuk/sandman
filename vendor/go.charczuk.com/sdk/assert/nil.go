package assert

import "reflect"

// Nil returns if a given reference is nil, but also returning true
// if the reference is a valid typed pointer to nil, which may not strictly
// be equal to nil.
func Nil(object any) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}
