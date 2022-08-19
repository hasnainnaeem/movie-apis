package main

type Config struct {
	AppPort   int    `mapstructure:"APP_PORT"`
	DBDriver  string `mapstructure:"DB_DRIVER"`
	DBHost    string `mapstructure:"DB_HOST"`
	DBPort    int    `mapstructure:"DB_PORT"`
	DBUser    string `mapstructure:"DB_USER"`
	DBPass    string `mapstructure:"DB_PASS"`
	DBName    string `mapstructure:"DB_NAME"`
	DBSSLMode string `mapstructure:"SSL_MODE"`

	RedisHost     string `mapstructure:"REDIS_URL"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
}
