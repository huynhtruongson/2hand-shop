package config

// ElasticsearchConfig holds the Elasticsearch connection parameters.
type ElasticsearchConfig struct {
	Address  string `mapstructure:"elasticsearch_address"` // ["http://localhost:9200"]
	Username string `mapstructure:"elasticsearch_username"`
	Password string `mapstructure:"elasticsearch_password"`
}
