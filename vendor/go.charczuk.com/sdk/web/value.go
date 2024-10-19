package web

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.charczuk.com/sdk/uuid"
)

// RouteValue parses a given route parameter value as retrieved from a given key.
func RouteValue[T any](ctx Context, key string) (out T, err error) {
	value, ok := ctx.RouteParams().Get(key)
	if !ok {
		err = fmt.Errorf("required route parameter missing; %s", key)
		return
	}
	err = ParseValue(&out, value)
	return
}

// QueryValue parses a given query string value as retrieved from a given key.
func QueryValue[T any](ctx Context, key string) (out T, err error) {
	value := ctx.Request().URL.Query().Get(key)
	if value == "" {
		err = fmt.Errorf("required query value missing; %s", key)
		return
	}
	err = ParseValue(&out, value)
	return
}

// HeaderValue parses a given header value as retrieved from a given key.
func HeaderValue[T any](ctx Context, key string) (out T, err error) {
	value := ctx.Request().Header.Get(key)
	if value == "" {
		err = fmt.Errorf("required header value missing; %s", key)
		return
	}
	err = ParseValue(&out, value)
	return
}

// FormValue parses a given post form value as retrieved from a given key.
func FormValue[T any](ctx Context, key string) (out T, err error) {
	value := ctx.Request().PostFormValue(key)
	if value == "" {
		err = fmt.Errorf("required post form value missing; %s", key)
		return
	}
	err = ParseValue(&out, value)
	return
}

// ParseValue is a helper to parse string values as given output values.
func ParseValue(value any, raw string) error {
	switch typed := value.(type) {
	case *bool:
		switch strings.ToLower(raw) {
		case "true":
			*typed = true
			return nil
		case "false":
			*typed = false
			return nil
		default:
			return fmt.Errorf("invalid boolean value: %v", raw)
		}
	case *string:
		*typed = raw
		return nil
	case *int:
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		*typed = parsed
		return nil
	case *int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return err
		}
		*typed = parsed
		return nil
	case *float32:
		parsed, err := strconv.ParseFloat(raw, 32)
		if err != nil {
			return err
		}
		*typed = float32(parsed)
		return nil
	case *float64:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		*typed = parsed
		return nil
	case *time.Duration:
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		*typed = parsed
		return nil
	case *uuid.UUID:
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return err
		}
		*typed = parsed
		return nil
	default:
		return errors.New("invalid parse value target")
	}
}
