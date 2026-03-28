package message

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

const ContentTypeJSON = "application/json"

type IMessage interface {
	EventName() string
	Exchange() string
	Message() RabbitMQMessage
}

type Envelope[T any] struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	Payload       T         `json:"payload"`
}

// NewEnvelope builds a ready-to-publish Envelope for the given payload.
func NewEnvelope[T any](msgType string, payload T, opts ...EnvelopeOption) *Envelope[T] {
	e := &Envelope[T]{
		ID:        uuid.New().String(),
		Type:      msgType,
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

// EnvelopeOption is a functional option for NewEnvelope.
type EnvelopeOption func(h any)

// WithCorrelationID sets the correlation ID on any *Envelope[T].
func WithCorrelationID(id string) EnvelopeOption {
	return func(h any) {
		// We use a type-switch so callers never need to worry about T.
		type correlatable interface{ setCorrelationID(string) }
		if c, ok := h.(correlatable); ok {
			c.setCorrelationID(id)
		}
	}
}

func (e *Envelope[T]) setCorrelationID(id string) { e.CorrelationID = id }

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
func NewRabbitMQMessage[T any](envelope *Envelope[T], opts ...MessageOption) (*RabbitMQMessage, error) {
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

// Decode unmarshals the delivery body into an Envelope[T].
// Returns an error if the body is not a valid JSON-encoded Envelope[T].
func Decode[T any](d *DeliveryMessage) (*Envelope[T], error) {
	var env Envelope[T]
	if err := json.Unmarshal(d.raw.Body, &env); err != nil {
		return nil, fmt.Errorf("message: decode envelope (type=%s): %w", d.raw.Type, err)
	}
	return &env, nil
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
