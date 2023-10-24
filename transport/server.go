package transport

import "context"

type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}
