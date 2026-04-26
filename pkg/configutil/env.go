package configutil

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// Env returns an environment variable with a given key as a value source.
func Env[T any](key string) func(context.Context) (*T, error) {
	return func(ctx context.Context) (*T, error) {
		output := new(T)
		raw, ok := GetEnvVars(ctx)[key]
		if !ok {
			return nil, nil
		}
		if err := parseGeneric(output, raw); err != nil {
			return nil, err
		}
		return output, nil
	}
}

func parseGeneric(ref any, raw string) error {
	switch refv := ref.(type) {
	case *string:
		*refv = raw
		return nil
	case *[]byte:
		*refv = []byte(raw)
		return nil
	case *float64:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		*refv = parsed
		return nil
	case *float32:
		parsed, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return err
		}
		*refv = float32(parsed)
		return nil
	case *int:
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		*refv = parsed
		return nil
	case *int16:
		parsed, err := strconv.ParseInt(raw, 10, 16)
		if err != nil {
			return err
		}
		*refv = int16(parsed)
		return nil
	case *int32:
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return err
		}
		*refv = int32(parsed)
		return nil
	case *int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		*refv = parsed
		return nil
	case *uint8:
		parsed, err := strconv.ParseUint(raw, 10, 8)
		if err != nil {
			return err
		}
		*refv = uint8(parsed)
		return nil
	case *uint16:
		parsed, err := strconv.ParseUint(raw, 10, 16)
		if err != nil {
			return err
		}
		*refv = uint16(parsed)
		return nil
	case *uint32:
		parsed, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return err
		}
		*refv = uint32(parsed)
		return nil
	case *uint64:
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return err
		}
		*refv = parsed
		return nil
	case *time.Duration:
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		*refv = parsed
		return nil
	default:
		return fmt.Errorf("invalid parse target: %T", ref)
	}
}
