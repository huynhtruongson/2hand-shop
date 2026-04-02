package dispatcher

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

const (
	// defaultRetryAttempts is the number of times a handler invocation is retried.
	defaultRetryAttempts = 3

	// defaultRetryDelay is the base delay between retry attempts.
	defaultRetryDelay = 300 * time.Millisecond
)

// DefaultRetryOptions contains the standard retry configuration used by the dispatcher.
// Handler errors are retried up to 3 times with an exponential-backoff delay of 300 ms.
var DefaultRetryOptions = []retry.Option{
	retry.Attempts(defaultRetryAttempts),
	retry.Delay(defaultRetryDelay),
	retry.DelayType(retry.BackOffDelay),
}

// EventDispatcher dispatches incoming RabbitMQ messages to registered typed handlers
// based on the routing key extracted from the message metadata.
//
// It implements consumer.RabbitMQConsumerHandler so it can be passed directly to
// RabbitMQConsumerConfiguration.Handler. The dispatcher is safe for concurrent use
// across multiple consumer goroutines.
type EventDispatcher struct {
	mu        sync.RWMutex
	registry  map[string]Handler // exact routing key → handler
	wildcards map[string]Handler // wildcard pattern → handler (checked after registry miss)
	log       logger.Logger
	retryOpts []retry.Option
}

// NewEventDispatcher constructs an EventDispatcher with the supplied logger and
// retry options. If retryOpts is nil, DefaultRetryOptions is used.
func NewEventDispatcher(log logger.Logger, retryOpts []retry.Option) *EventDispatcher {
	opts := retryOpts
	if opts == nil {
		opts = DefaultRetryOptions
	}
	return &EventDispatcher{
		registry:  make(map[string]Handler),
		wildcards: make(map[string]Handler),
		log:       log,
		retryOpts: opts,
	}
}

// Handle implements consumer.RabbitMQConsumerHandler. It routes the incoming message
// to the handler that best matches its routing key, falling back through the following
// lookup order:
//
//  1. Exact routing key match (e.g. "catalog.product.created")
//  2. Wildcard pattern match (e.g. "catalog.product.*")
//
// If no handler is found the message is acknowledged with a warning log and no error
// is returned to the consumer, preventing unnecessary redelivery of unhandled messages.
// If a handler is found but fails, the retry policy is applied before returning the error.
func (d *EventDispatcher) Handle(ctx context.Context, msg *types.DeliveryMessage) error {
	meta := msg.Metadata()
	// Use RoutingKey (AMQP-level) so routing works even when the JSON body is
	// malformed and meta.Type cannot be unmarshalled. meta.Type (envelope type)
	// is still logged for visibility.
	routingKey := meta.RoutingKey

	handler, ok := d.lookupHandler(routingKey)
	if !ok {
		d.log.Warn("dispatcher: no handler registered for routing key, acknowledging message",
			"routing_key", routingKey,
			"exchange", meta.Exchange,
			"delivery_tag", meta.DeliveryTag,
		)
		return nil
	}

	var lastErr error
	err := retry.Do(
		func() error {
			lastErr = handler.Handle(ctx, msg)
			return lastErr
		},
		append(d.retryOpts, retry.Context(ctx))...,
	)
	if err != nil {
		d.log.Error("dispatcher: handler failed after retries",
			"routing_key", routingKey,
			"delivery_tag", meta.DeliveryTag,
			"error", err,
		)
		return err
	}

	return nil
}

// lookupHandler performs a thread-safe lookup of the handler for routingKey.
// It first checks for an exact match, then checks wildcard patterns.
// The returned boolean indicates whether a handler was found.
func (d *EventDispatcher) lookupHandler(routingKey string) (Handler, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if h, ok := d.registry[routingKey]; ok {
		return h, true
	}

	for pattern, h := range d.wildcards {
		if patternMatch(routingKey, pattern) {
			return h, true
		}
	}

	return nil, false
}

// register adds a handler to the dispatcher's registry under the given routing key.
// It is called exclusively during Build() while holding the write lock.
func (d *EventDispatcher) register(routingKey string, handler Handler) {
	d.registry[routingKey] = handler
}

// registerWildcard adds a handler to the wildcard registry.
// Wildcard handlers are checked after exact-match lookups fail.
// It is called exclusively during Build() while holding the write lock.
func (d *EventDispatcher) registerWildcard(pattern string, handler Handler) {
	d.wildcards[pattern] = handler
}

// patternMatch returns true if routingKey matches the given pattern.
// A single asterisk (*) in a pattern segment matches any non-empty string
// in the corresponding routing key segment. Patterns must contain the same
// number of dot-separated segments as the routing key.
//
// Examples:
//
//	patternMatch("catalog.product.created", "catalog.product.*")   → true
//	patternMatch("catalog.product.created", "catalog.product.created") → true
//	patternMatch("catalog.product.created", "catalog.*.created")  → true
//	patternMatch("catalog.product.created", "catalog.order.*")     → false
//	patternMatch("catalog.product.created", "catalog.product")     → false
func patternMatch(routingKey, pattern string) bool {
	rkParts := strings.Split(routingKey, ".")
	patParts := strings.Split(pattern, ".")

	if len(rkParts) != len(patParts) {
		return false
	}

	for i := range rkParts {
		if patParts[i] != "*" && patParts[i] != rkParts[i] {
			return false
		}
	}

	return true
}
