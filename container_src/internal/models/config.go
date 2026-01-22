package models

type Config struct{}

func NewConfig() *Config {
	return new(Config)
}
