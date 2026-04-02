package types

import (
	"time"
)

// EventContext is the context passed to every event handler registered with
// the dispatcher. It exposes the decoded domain-event payload alongside its
// full envelope metadata (correlation ID, timestamp, RabbitMQ delivery info).
//
// Handlers receive this interface so they can access both the payload and
// the metadata without the dispatcher needing to know the concrete event type
// at the point of registration.
type EventContext interface {
	// Payload returns the fully-decoded domain event. Callers type-assert:
	//
	//	p, ok := ec.Payload().(ProductCreatedEvent)
	//	if !ok { return nil }
	Payload() any

	// ID returns the domain event UUID.
	ID() string
	// Type returns the event type string, e.g. "product.created".
	Type() string
	// Timestamp returns when the event was originally emitted by the producer.
	Timestamp() time.Time
	// CorrelationID returns the distributed-tracing correlation ID, if set.
	CorrelationID() string
	// Metadata returns the raw RabbitMQ delivery metadata (exchange, routing key, delivery tag).
	Metadata() Metadata
}

// eventContext is the internal concrete type that satisfies EventContext.
type eventContext struct {
	envelope EventEnvelope
	meta     Metadata
}

func NewEventContext(env EventEnvelope, meta Metadata) EventContext {
	return &eventContext{envelope: env, meta: meta}
}

func (e *eventContext) Payload() any          { return e.envelope.Payload }
func (e *eventContext) ID() string            { return e.envelope.ID }
func (e *eventContext) Type() string          { return e.envelope.Type }
func (e *eventContext) Timestamp() time.Time  { return e.envelope.Timestamp }
func (e *eventContext) CorrelationID() string { return e.envelope.CorrelationID }
func (e *eventContext) Metadata() Metadata    { return e.meta }
