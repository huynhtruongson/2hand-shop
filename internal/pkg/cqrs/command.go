package cqrs

import "context"

type CommandHandler[TCommand any, TResponse any] interface {
	Handle(ctx context.Context, command TCommand) (TResponse, error)
}
