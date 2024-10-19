package configutil

import "context"

// Set sets a field from a given list of sources.
func Set[T any](destination *T, sources ...Source[T]) ResolveAction {
	return func(ctx context.Context) error {
		var value *T
		var err error
		for _, source := range sources {
			value, err = source(ctx)
			if err != nil {
				return err
			}
			if value != nil {
				*destination = *value
				return nil
			}
		}
		return nil
	}
}
