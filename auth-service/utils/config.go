package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DSN                     string        `mapstructure:"DSN"`
	REDIS_PASSWORD          string        `mapstructure:"REDIS_PASSWORD"`
	REDIS_HOST              string        `mapstructure:"REDIS_HOST"`
	REDIS_PORT              string        `mapstructure:"REDIS_PORT"`
	MIGRATION_URL           string        `mapstructure:"MIGRATION_URL"`
	DB_MAX_OPEN_CONNECTION  int           `mapstructure:"DB_MAX_OPEN_CONNECTION"`
	DB_MAX_IDLE_CONNECTION  int           `mapstructure:"DB_MAX_IDLE_CONNECTION"`
	DB_MAX_IDLE_TIME        string        `mapstructure:"DB_MAX_IDLE_TIME"`
	PORT                    int           `mapstructure:"PORT"`
	TOKEN_SYMMETRIC_KEY     string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	ACCESS_TOKEN_DURATION   time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	REFRESH_TOKEN_DURATION  time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	ACTIVATE_TOKEN_DURATION time.Duration `mapstructure:"ACTIVATE_TOKEN_DURATION"`
	AUTH_EMAIL              string        `mapstructure:"AUTH_EMAIL"`
}

func LoadConfig(path string) (config Config, err error) {

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
