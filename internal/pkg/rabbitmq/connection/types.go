package connection

type ConnectionType string

var (
	ConnectionTypeConsumer ConnectionType = "consumer"
	ConnectionTypeProducer ConnectionType = "producer"
)
