package cqrs

import "context"

type QueryHandler[TQuery any, TResponse any] interface {
	Handle(ctx context.Context, query TQuery) (TResponse, error)
}
