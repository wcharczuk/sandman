package assert

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"
)

// AreEqual is a helper that returns if two references are equal.
func AreEqual(expected, actual any) bool {
	expectedIsNil := ReferenceIsNil(expected)
	actualIsNil := ReferenceIsNil(actual)
	if expectedIsNil && actualIsNil {
		return true
	}
	if (expectedIsNil && !actualIsNil) || (!expectedIsNil && actualIsNil) {
		return false
	}
	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}
	return reflect.DeepEqual(expected, actual)
}

// InTimeDelta returns if two times are within a delta as a duration.
func InTimeDelta(a, b time.Time, delta time.Duration) bool {
	actual := a.Sub(b)
	if actual < 0 {
		return -actual < delta
	}
	return actual < delta
}

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

// ReferenceIsNil returns if a given reference is nil, but also returning true
// if the reference is a valid typed pointer to nil, which may not strictly
// be equal to nil.
func ReferenceIsNil(object any) bool {
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

// Fatalf is a helper to fatal a test with optional message components.
func Fatalf(t *testing.T, format string, args, message []any) {
	t.Helper()
	if len(message) > 0 {
		t.Fatal(fmt.Sprintf(format, args...) + ": " + fmt.Sprint(message...))
	} else {
		t.Fatalf(format, args...)
	}
}

// StringContains returns if a substr from an any reference is contained in a given string.
func StringContains(s string, substr any) bool {
	switch typed := substr.(type) {
	case string:
		return strings.Contains(s, typed)
	case *string:
		if typed != nil {
			return strings.Contains(s, *typed)
		}
		return false
	default:
		return strings.Contains(s, fmt.Sprint(typed))
	}
}

// StringHasPrefix is a helper that returns if a string has a given prefix from an any reference.
func StringHasPrefix(corpus string, prefix any) bool {
	switch typed := prefix.(type) {
	case string:
		return strings.HasPrefix(corpus, typed)
	case *string:
		if typed != nil {
			return strings.HasPrefix(corpus, *typed)
		}
		return false
	default:
		return strings.HasPrefix(corpus, fmt.Sprint(typed))
	}
}

// StringHasSuffix is a helper that returns if a string has a given suffix from an any reference.
func StringHasSuffix(corpus string, suffix any) bool {
	switch typed := suffix.(type) {
	case string:
		return strings.HasSuffix(corpus, typed)
	case *string:
		if typed != nil {
			return strings.HasSuffix(corpus, *typed)
		}
		return false
	default:
		return strings.HasSuffix(corpus, fmt.Sprint(typed))
	}
}

// RegexpMatches returns if a given string compiled as a regexp matches the actual.
func RegexpMatches(expr string, actual any) bool {
	compiled := regexp.MustCompile(expr)
	switch typed := actual.(type) {
	case string:
		return compiled.MatchString(typed)
	case *string:
		if typed == nil {
			return false
		}
		return compiled.MatchString(*typed)
	default:
		return compiled.MatchString(fmt.Sprint(actual))
	}
}
