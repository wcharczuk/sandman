package selector

import (
	"fmt"
)

// CheckLabels validates all the keys and values for the label set.
func CheckLabels(labels Labels) (err error) {
	for key, value := range labels {
		err = CheckKey(key)
		if err != nil {
			err = fmt.Errorf("invalid label key %s; %w", key, err)
			return
		}
		err = CheckValue(value)
		if err != nil {
			err = fmt.Errorf("invalid label key %s and value %s; %w", key, value, err)
			return
		}
	}
	return
}
