package assert

import (
	"fmt"
	"testing"
)

// Fatalf is a helper to fatal a test with optional message components.
func Fatalf(t *testing.T, format string, args, message []any) {
	t.Helper()
	if len(message) > 0 {
		t.Fatal(fmt.Sprintf(format, args...) + ": " + fmt.Sprint(message...))
	} else {
		t.Fatalf(format, args...)
	}
}
