package types

type DomainEvent interface {
	// EventType is a dot-separated descriptor used as the Envelope.Type
	// and the AMQP Publishing.Type, e.g. "order.created".
	EventType() string

	// Exchange is the AMQP exchange to publish to.
	Exchange() string

	// CorrelationID links this event to a parent request or saga.
	// Return an empty string if not applicable.
	CorrelationID() string
}

type BaseEvent struct {
	eventType     string
	exchange      string
	correlationID string
}

func NewBaseEvent(eventType, exchange, correlationID string) BaseEvent {
	return BaseEvent{
		eventType:     eventType,
		exchange:      exchange,
		correlationID: correlationID,
	}
}

func (b BaseEvent) EventType() string     { return b.eventType }
func (b BaseEvent) Exchange() string      { return b.exchange }
func (b BaseEvent) CorrelationID() string { return b.correlationID }
