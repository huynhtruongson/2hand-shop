package config

type CognitoConfig struct {
	Region       string `mapstructure:"cognito_region"`
	UserPoolID   string `mapstructure:"cognito_user_pool_id"`
	ClientID     string `mapstructure:"cognito_client_id"`
	ClientSecret string `mapstructure:"cognito_client_secret"`
}
