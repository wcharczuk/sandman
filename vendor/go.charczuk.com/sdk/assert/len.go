package assert

import "reflect"

// Len returns the length of a given reference if it is a slice, string, channel, or map.
func Len(object any) int {
	switch object {
	case nil:
		return 0
	case "":
		return 0
	}
	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	case reflect.Map, reflect.Slice, reflect.Chan, reflect.String:
		return objValue.Len()
	default:
		return 0
	}
}
