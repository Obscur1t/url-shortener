package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App App
	DB  PostgresDB
}

type App struct {
	AppEnv  string `env:"APP_ENV" env-default:"local"`
	AppAddr string `env:"APP_ADDR" env-default:":8080"`
}

type PostgresDB struct {
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
