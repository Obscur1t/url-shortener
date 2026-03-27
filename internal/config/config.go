package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App  AppConfig
	DB   PostgresDBConfig
	Auth AuthConfig
}

type AppConfig struct {
	AppEnv  string `env:"APP_ENV" env-default:"local"`
	AppAddr string `env:"APP_ADDR" env-default:":8080"`
}

type AuthConfig struct {
	Secret   string        `env:"Auth_Secret" env-required:"true"`
	TokenTTL time.Duration `env:"Auth_TokenTTL" env-required:"true"`
}

type PostgresDBConfig struct {
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	Host     string `env:"POSTGRES_HOST" env-default:"postgres"`
	SSL      string `env:"POSTGRES_SSL" env-default:"disable"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg := Config{}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}
