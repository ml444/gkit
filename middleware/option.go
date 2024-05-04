package middleware

import "context"

// LurkerFunc  The interface of the middleware processing function
// allows the user to implement the processing logic himself.
type LurkerFunc func(ctx context.Context, p interface{}) (err error)

func LurkerChain(ctx context.Context, p interface{}, fns ...LurkerFunc) error {
	for _, fn := range fns {
		err := fn(ctx, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func ForceLurkerChain(ctx context.Context, p interface{}, fns ...LurkerFunc) {
	for _, fn := range fns {
		_ = fn(ctx, p)
	}
}
