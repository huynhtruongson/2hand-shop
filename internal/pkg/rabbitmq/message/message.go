package message

type RabbitMQMessage struct {
	Data              interface{}
	PublishingBuilder RabbitMQPublishingBuilder
}
type RabbitMQPublishingBuilderBuilderFunc func(builder RabbitMQPublishingBuilder)

func NewRabbitMQMessage(data interface{}, publishingBuilderFn RabbitMQPublishingBuilderBuilderFunc) *RabbitMQMessage {
	builder := NewRabbitMQPublishingBuilder()
	if publishingBuilderFn != nil {
		publishingBuilderFn(builder)
	}
	return &RabbitMQMessage{
		Data:              data,
		PublishingBuilder: builder,
	}
}
