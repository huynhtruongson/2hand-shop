package types

import (
	"time"
)

type EventContext[T any] interface {
	Payload() T
	ID() string
	Type() string
	Timestamp() time.Time
	CorrelationID() string
	Metadata() Metadata
}

type eventContext[T any] struct {
	envelope EventEnvelope[T]
	meta     Metadata
}

func NewEventContext[T any](env EventEnvelope[T], meta Metadata) EventContext[T] {
	return &eventContext[T]{envelope: env, meta: meta}
}

func (e *eventContext[T]) Payload() T          { return e.envelope.Payload }
func (e *eventContext[T]) ID() string           { return e.envelope.ID }
func (e *eventContext[T]) Type() string          { return e.envelope.Type }
func (e *eventContext[T]) Timestamp() time.Time  { return e.envelope.Timestamp }
func (e *eventContext[T]) CorrelationID() string { return e.envelope.CorrelationID }
func (e *eventContext[T]) Metadata() Metadata   { return e.meta }
