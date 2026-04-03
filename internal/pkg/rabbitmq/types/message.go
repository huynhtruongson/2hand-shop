package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

const ContentTypeJSON = "application/json"

type EventEnvelope[T any] struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Payload       T         `json:"payload"`
}

// NewEnvelope builds a ready-to-publish EventEnvelope for the given payload.
func NewEventEnvelope(msgType string, payload any, correlationID string) *EventEnvelope[any] {
	return &EventEnvelope[any]{
		ID:            uuid.New().String(),
		Type:          msgType,
		Timestamp:     time.Now().UTC(),
		Payload:       payload,
		CorrelationID: correlationID,
	}
}

// --------------------------------- publishing ---------------------------------

// RabbitMQMessage is what the producer accepts.
// It bundles the serialised envelope with AMQP-level publishing options.
type RabbitMQMessage struct {
	// Type is copied into amqp091.Publishing.Type for broker-level routing.
	Type string
	// CorrelationID is copied into amqp091.Publishing.CorrelationId.
	CorrelationID string
	// Body is the JSON-encoded Envelope. Set by NewRabbitMQMessage.
	Body []byte
	// Headers are optional AMQP headers merged on top of defaults.
	Headers amqp091.Table
}

// NewRabbitMQMessage serialises envelope into a RabbitMQMessage ready to hand
// to the producer. T must be JSON-serialisable.
func NewRabbitMQMessage[T any](envelope *EventEnvelope[T], opts ...MessageOption) (*RabbitMQMessage, error) {
	body, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("message: marshal envelope: %w", err)
	}
	m := &RabbitMQMessage{
		Type:          envelope.Type,
		CorrelationID: envelope.CorrelationID,
		Body:          body,
	}
	for _, o := range opts {
		o(m)
	}
	return m, nil
}

// MessageOption is a functional option for NewRabbitMQMessage.
type MessageOption func(*RabbitMQMessage)

// WithHeaders merges extra AMQP headers into the message.
func WithHeaders(h amqp091.Table) MessageOption {
	return func(m *RabbitMQMessage) {
		if m.Headers == nil {
			m.Headers = make(amqp091.Table)
		}
		for k, v := range h {
			m.Headers[k] = v
		}
	}
}

// ToPublishing converts the RabbitMQMessage to an amqp091.Publishing.
// Called internally by the producer.
func (m *RabbitMQMessage) ToPublishing(appID string) amqp091.Publishing {
	return amqp091.Publishing{
		ContentType:   ContentTypeJSON,
		Type:          m.Type,
		CorrelationId: m.CorrelationID,
		MessageId:     uuid.New().String(),
		Timestamp:     time.Now().UTC(),
		AppId:         appID,
		Body:          m.Body,
		Headers:       m.Headers,
		DeliveryMode:  amqp091.Persistent,
	}
}

// --------------------------------- consuming ----------------------------------

// DeliveryMessage wraps an amqp091.Delivery and exposes typed unmarshalling.
// Consumers receive this instead of the raw delivery.
type DeliveryMessage struct {
	raw *amqp091.Delivery
}

// NewDeliveryMessage wraps a raw AMQP delivery.
func NewDeliveryMessage(d *amqp091.Delivery) *DeliveryMessage {
	return &DeliveryMessage{raw: d}
}

// DecodeEnvelope decodes the RabbitMQ delivery body into an EventEnvelope[T]
// and returns it alongside the AMQP-level Metadata (exchange, routing key, delivery tag).
// Returns an error if the body is not valid JSON or cannot be assigned to EventEnvelope[T].
func DecodeEnvelope[T any](d *DeliveryMessage) (EventEnvelope[T], Metadata, error) {
	var env EventEnvelope[T]
	if err := json.Unmarshal(d.raw.Body, &env); err != nil {
		return EventEnvelope[T]{}, Metadata{}, fmt.Errorf("message: decode envelope (type=%s): %w", d.raw.Type, err)
	}
	return env, d.Metadata(), nil
}

// Metadata returns envelope-level fields without fully decoding the payload.
// Useful for logging/routing without paying the cost of generic unmarshalling.
func (d *DeliveryMessage) Metadata() Metadata {
	var meta Metadata
	// Best-effort; ignore decode errors — Metadata is advisory only.
	_ = json.Unmarshal(d.raw.Body, &meta)
	meta.Exchange = d.raw.Exchange
	meta.RoutingKey = d.raw.RoutingKey
	meta.DeliveryTag = d.raw.DeliveryTag
	return meta
}

// Metadata holds the fields that can be read without knowing the payload type.
type Metadata struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id"`
	// AMQP-level fields (not in JSON body)
	Exchange    string `json:"-"`
	RoutingKey  string `json:"-"`
	DeliveryTag uint64 `json:"-"`
}

// Raw returns the underlying amqp091.Delivery for cases where direct access
// is needed (e.g. reading custom AMQP headers).
func (d *DeliveryMessage) Raw() *amqp091.Delivery { return d.raw }
