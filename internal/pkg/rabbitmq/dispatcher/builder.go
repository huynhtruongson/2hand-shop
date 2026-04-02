package dispatcher

import (
	"errors"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

// ErrEmptyRoutingKey is returned by Builder.Build when a handler is registered
// with an empty routing key or pattern.
var ErrEmptyRoutingKey = errors.New("dispatcher: routing key must not be empty")

// ErrNilHandler is returned by Builder.Build when a nil handler function is registered.
var ErrNilHandler = errors.New("dispatcher: handler function must not be nil")

// Builder provides a fluent API for constructing an EventDispatcher.
// Register handlers using Register (exact match) or RegisterWildcard (wildcard
// pattern matching), then call Build to produce the final dispatcher.
//
// Builder is not safe for concurrent use; register all handlers before calling Build.
type Builder struct {
	log       logger.Logger
	handlers  map[string]Handler
	wildcards map[string]Handler
	retryOpts []retry.Option
}

// NewBuilder creates a new Builder instance.
func NewBuilder(log logger.Logger) *Builder {
	return &Builder{
		log:       log,
		handlers:  make(map[string]Handler),
		wildcards: make(map[string]Handler),
	}
}

// WithRetryOptions sets the retry options used by the dispatcher.
// By default, DefaultRetryOptions (3 attempts, 300 ms backoff) is used.
// This method must be called before Register or RegisterWildcard if used.
func (b *Builder) WithRetryOptions(opts ...retry.Option) *Builder {
	b.retryOpts = opts
	return b
}

// Register registers fn as the handler for the exact routing key routingKey.
// fn is an EventHandler — a plain function that accepts (ctx, EventContext).
// The generic T binds the JSON decode target so TypedPayload[T]() is type-safe.
//
// Example:
//
//	dispatcher.Register[ProductCreatedEvent](b, "catalog.product.created",
//	    func(ctx context.Context, ec EventContext) error {
//	        return nil
//	    },
//	)
func Register(b *Builder, routingKey string, fn types.EventHandler) *Builder {
	if err := validateKey(routingKey); err != nil {
		panic(err)
	}
	if fn == nil {
		panic(ErrNilHandler)
	}
	b.handlers[routingKey] = NewEventHandler(fn)
	return b
}

// RegisterWildcard registers fn as the handler for any routing key that matches
// the wildcard pattern. The pattern uses a single asterisk (*) per segment to
// match any non-empty string in that segment (e.g. "catalog.product.*" matches
// both "catalog.product.created" and "catalog.product.deleted").
//
// Wildcard handlers are only consulted when no exact-match handler is found.
//
// Example:
//
//	dispatcher.RegisterWildcard[any](b, "catalog.product.*",
//	    func(ctx context.Context, ec EventContext) error {
//	        return nil
//	    },
//	)
func RegisterWildcard(b *Builder, pattern string, fn types.EventHandler) *Builder {
	if err := validateKey(pattern); err != nil {
		panic(err)
	}
	if fn == nil {
		panic(ErrNilHandler)
	}
	if !strings.Contains(pattern, "*") {
		panic(fmt.Errorf("dispatcher: pattern %q contains no wildcard character '*'", pattern))
	}
	b.wildcards[pattern] = NewEventHandler(fn)
	return b
}

// Build validates the registered handlers and produces an EventDispatcher ready
// to be passed to a RabbitMQ consumer configuration.
//
// It panics if any registration is invalid (empty key, nil handler, pattern without '*').
// Build is typically called once during application startup.
func (b *Builder) Build() (*EventDispatcher, error) {
	if len(b.handlers) == 0 && len(b.wildcards) == 0 {
		return nil, errors.New("dispatcher: no handlers registered")
	}

	dispatcher := NewEventDispatcher(b.log, b.retryOpts)

	for k, h := range b.handlers {
		dispatcher.register(k, h)
	}
	for p, h := range b.wildcards {
		dispatcher.registerWildcard(p, h)
	}

	return dispatcher, nil
}

// validateKey checks that a routing key or pattern is non-empty and contains at least
// one dot-separated segment.
func validateKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrEmptyRoutingKey
	}
	if !strings.Contains(key, ".") {
		return fmt.Errorf("dispatcher: routing key %q must contain at least one dot-separated segment", key)
	}
	return nil
}
