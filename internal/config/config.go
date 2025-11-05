package config

import "github.com/spf13/viper"

func Load() error {
	viper.SetDefault("API_ADDR", ":8080")
	viper.SetDefault("DB_DSN", "postgres://postgres:postgres@localhost:5432/energy?sslmode=disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("MQTT_BROKER", "tcp://localhost:1883")
	viper.AutomaticEnv()
	return nil
}

func MQTTBroker() string { return viper.GetString("MQTT_BROKER") }
