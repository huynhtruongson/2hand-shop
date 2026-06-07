package config

type KeycloakConfig struct {
	Realm        string `mapstructure:"keycloak_realm"`
	BaseURL      string `mapstructure:"keycloak_base_url"`
	ClientID     string `mapstructure:"keycloak_client_id"`
	ClientSecret string `mapstructure:"keycloak_client_secret"`
}
