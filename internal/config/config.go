package config

import (
	"avito/pkg"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string `env:"PORT" env-default:"8080"`

	MaxAttempts int `env:"MAX_ATTEMPTS" env-default:"1"`
	Delay       int `env:"DELAY" env-default:"1"`

	PeriodCheck   int `env:"PERIOD_CHECK" env-default:"10"`
	CacheTTL      int `env:"CACHE_TTL" env-default:"60"`
	CacheSize     int `env:"CACHE_SIZE" env-default:"1"`
	BatchSize     int `env:"BATCH_SIZE" env-default:"3"`
	PollingPeriod int `env:"POLLING_PERIOD" env-default:"5"`

	OrderServiceAddress     string `env:"ORDER_SERVICE_HOST" env-required:"true"`
	ConnRetryToOrderService int    `env:"RETRY_CONNECT_TO_ORDER_SERVICE" env-default:"3"`
	ConnectDelay            int    `env:"CONNECT_DELAY" env-default:"1"`
	RetryDelay              int    `env:"RETRY_DELAY" env-default:"100"`

	KafkaTopic string `env:"KAFKA_TOPIC" env-default:"test-topic"`
	KafkaHost  string `env:"KAFKA_HOST" env-required:"true"`
	KafkaPort  string `env:"KAFKA_PORT" env-default:"9092"`

	IpTTL      int     `env:"IP_TTL" env-default:"10"`
	MaxTokens  float64 `env:"MAX_TOKENS" env-default:"10"`
	RefillRate float64 `env:"REFILL_RATE" env-default:"5"`

	EnableProfiling bool   `env:"ENABLE_PROFILING" env-default:"false"`
	ProfilePort     string `env:"PROFILE_PORT" env-default:"6060"`

	Login    string `env:"BASE_LOGIN" env-default:"admin"`
	Password string `env:"BASE_PASSWORD" env-default:"admin"`

	logger pkg.Logger `env:"-"`
}

func NewConfig(logger pkg.Logger) *Config {
	return &Config{logger: logger}
}

func (c *Config) SetCFGFromEnvFile() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		_ = godotenv.Load(p)
	}
}

func (c *Config) LoadConfig() {
	if err := cleanenv.ReadEnv(c); err != nil {
		c.logger.Error("config read error: %v", err)
		panic(err)
	}

	c.KafkaHost = c.KafkaHost + ":" + c.KafkaPort

	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 1
	}
	if c.Delay <= 0 {
		c.Delay = 1
	}

	if c.ProfilePort == "" {
		c.ProfilePort = "6060"
	}
}
