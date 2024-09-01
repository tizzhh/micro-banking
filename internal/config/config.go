package config

import (
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-default:"1h"`
	GRPC        GRPCConfig    `yaml:"grpc" env-required:"true"`
	DB          DBConfig      `yaml:"db" env-required:"true"`
	Redis       RedisConfig   `yaml:"redis" env-required:"true"`
	CurrencyApi CurrencyApi   `yaml:"currency_api" env-required:"true"`
	Mail        Mail          `yaml:"mail" env-required:"true"`
	Clients     Clients       `yaml:"clients" env-required:"true"`
	Http        Http          `yaml:"http" env-required:"true"`
}

type Clients struct {
	AuthClient     AuthClient     `yaml:"auth" env-required:"true"`
	CurrencyClient CurrencyClient `yaml:"currency" env-required:"true"`
}

type AuthClient struct {
	Addr         string        `yaml:"addr" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-default:"10s"`
	RetriesCount int           `yaml:"retries_count" env-default:"3"`
}

type CurrencyClient struct {
	Addr         string        `yaml:"addr" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-default:"10s"`
	RetriesCount int           `yaml:"retries_count" env-default:"3"`
}

type Http struct {
	Port            int           `yaml:"port" env-required:"true"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"5s"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"3s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"3s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"3s"`
}

type Mail struct {
	From     string `yaml:"from" env-required:"true"`
	SmtpHost string `yaml:"smtp_host" env-required:"true"`
	SmtpPort int    `yaml:"smtp_port" env-default:"587"`
	ApiKey   string `yaml:"api_key" env-required:"true"`
}

type CurrencyApi struct {
	URL     string        `yaml:"url" env-required:"true"`
	ApiKey  string        `yaml:"api_key" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"3s"`
}

type GRPCConfig struct {
	AuthPort     int           `yaml:"auth_port" env-required:"true"`
	CurrencyPort int           `yaml:"currency_port" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-default:"10s"`
}

type DBConfig struct {
	DBName     string `yaml:"db_name" env-default:"micro_bank"`
	DBUser     string `yaml:"db_user" env-required:"true"`
	DBPassword string `yaml:"db_password" env-required:"true"`
	DBHost     string `yaml:"db_host" env-required:"true"`
	DBPort     int    `yaml:"db_port" env-default:"5432"`
}

type RedisConfig struct {
	Port        int           `yaml:"port" env-default:"6379"`
	Host        string        `yaml:"host" env-default:"localhost"`
	Password    string        `yaml:"password" env-required:"true"`
	PingTimeout time.Duration `yaml:"ping_timeout" env-default:"5s"`
	KeyTTL      time.Duration `yaml:"key_ttl" env-default:"1m"`
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
