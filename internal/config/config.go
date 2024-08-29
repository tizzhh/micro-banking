package config

import (
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string        `yaml:"env" env-required:"true"`
	TokenTTL time.Duration `yaml:"token_ttl" env-default:"1h"`
	GRPC     GRPCConfig    `yaml:"grpc"`
	DB       DBConfig      `yaml:"db"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

type DBConfig struct {
	DBName     string `yaml:"db_name" env-default:"micro_bank"`
	DBUser     string `yaml:"db_user" env-required:"true"`
	DBPassword string `yaml:"db_password" env-required:"true"`
	DBHost     string `yaml:"db_host" env-required:"true"`
	DBPort     int    `yaml:"db_port" env-default:"5432"`
}

var configSingleton *Config
var once sync.Once

func Get() *Config {
	once.Do(func() {
		configSingleton = MustLoad()
	})
	return configSingleton
}

func MustLoad() *Config {
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		panic("CONFIG_PATH is not set")
	}
	if configPath == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(configPath); err != nil {
		panic("config file does not exist")
	}

	_, ok = os.LookupEnv("SECRET_KEY")
	if !ok {
		panic("SECRET_KEY for JWT is not set")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}
