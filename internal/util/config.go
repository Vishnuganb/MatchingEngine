package util

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	DBSource    string `mapstructure:"DB_SOURCE"`
	DBDriver    string `mapstructure:"DB_DRIVER"`
	KafkaBroker string `mapstructure:"KAFKA_BROKER"`
	KafkaTopic  string `mapstructure:"KAFKA_TOPIC"`
	RmqHost     string `mapstructure:"RMQ_URL"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)     // directory path to search for the config file
	viper.SetConfigName("common") // looks for app.env
	viper.SetConfigType("env")    // // tells Viper to treat it like key=value format

	viper.AutomaticEnv()       // So this line makes it possible to override .env values using system environment variables — useful in Docker, CI/CD, etc.
	err = viper.ReadInConfig() // Try to read the configuration file — like app.env — from the path I gave earlier.
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config) // // maps values from config to struct fields using `mapstructure` tags
	return
}
