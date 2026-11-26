package config_test

import (
	"os"
	"testing"

	"avito/internal/config"
	"avito/pkg"
)

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	log := pkg.NewLogger(pkg.INFO)
	return config.NewConfig(log)
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}

func mustSetenv(t *testing.T, key, val string) {
	t.Helper()
	if err := os.Setenv(key, val); err != nil {
		t.Fatalf("setenv %s: %v", key, err)
	}
}

func mustLoad(t *testing.T, cfg *config.Config) {
	t.Helper()
	cfg.LoadConfig()
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic, got none")
		}
	}()
	fn()
}

func TestLoadConfig_DefaultPort(t *testing.T) {
	unsetEnv(t, "PORT", "ORDER_SERVICE_HOST", "KAFKA_HOST")

	mustSetenv(t, "ORDER_SERVICE_HOST", "http://order")
	mustSetenv(t, "KAFKA_HOST", "localhost")

	cfg := newTestConfig(t)
	mustLoad(t, cfg)

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
}

func TestLoadConfig_PortFromEnv(t *testing.T) {
	unsetEnv(t, "PORT", "ORDER_SERVICE_HOST", "KAFKA_HOST")

	mustSetenv(t, "PORT", "9090")
	mustSetenv(t, "ORDER_SERVICE_HOST", "http://order")
	mustSetenv(t, "KAFKA_HOST", "localhost")

	cfg := newTestConfig(t)
	mustLoad(t, cfg)

	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", cfg.Port)
	}
}

func TestLoadConfig_DefaultInts(t *testing.T) {
	unsetEnv(
		t,
		"MAX_ATTEMPTS",
		"DELAY",
		"PERIOD_CHECK",
		"CACHE_TTL",
		"CACHE_SIZE",
		"BATCH_SIZE",
		"RETRY_CONNECT_TO_ORDER_SERVICE",
		"CONNECT_DELAY",
		"POLLING_PERIOD",
		"ORDER_SERVICE_HOST",
		"KAFKA_HOST",
		"KAFKA_PORT",
		"KAFKA_TOPIC",
	)

	mustSetenv(t, "ORDER_SERVICE_HOST", "http://order")
	mustSetenv(t, "KAFKA_HOST", "localhost")

	cfg := newTestConfig(t)
	mustLoad(t, cfg)

	if cfg.MaxAttempts != 1 {
		t.Errorf("MaxAttempts expected 1, got %d", cfg.MaxAttempts)
	}
	if cfg.Delay != 1 {
		t.Errorf("Delay expected 1, got %d", cfg.Delay)
	}
	if cfg.PeriodCheck != 10 {
		t.Errorf("PeriodCheck expected 10, got %d", cfg.PeriodCheck)
	}
	if cfg.CacheTTL != 60 {
		t.Errorf("CacheTTL expected 60, got %d", cfg.CacheTTL)
	}
	if cfg.CacheSize != 1 {
		t.Errorf("CacheSize expected 1, got %d", cfg.CacheSize)
	}
	if cfg.BatchSize != 3 {
		t.Errorf("BatchSize expected 3, got %d", cfg.BatchSize)
	}
	if cfg.ConnRetryToOrderService != 3 {
		t.Errorf("ConnRetryToOrderService expected 3, got %d", cfg.ConnRetryToOrderService)
	}
	if cfg.ConnectDelay != 1 {
		t.Errorf("ConnectDelay expected 1, got %d", cfg.ConnectDelay)
	}
	if cfg.PollingPeriod != 5 {
		t.Errorf("PollingPeriod expected 5, got %d", cfg.PollingPeriod)
	}
}

func TestLoadConfig_KafkaDefaults(t *testing.T) {
	unsetEnv(
		t,
		"KAFKA_TOPIC",
		"KAFKA_HOST",
		"KAFKA_PORT",
		"ORDER_SERVICE_HOST",
	)

	mustSetenv(t, "ORDER_SERVICE_HOST", "http://order")
	mustSetenv(t, "KAFKA_HOST", "localhost")

	cfg := newTestConfig(t)
	mustLoad(t, cfg)

	if cfg.KafkaTopic != "test-topic" {
		t.Fatalf("expected default kafka topic, got %q", cfg.KafkaTopic)
	}

	if cfg.KafkaPort != "9092" {
		t.Fatalf("expected default kafka port 9092, got %q", cfg.KafkaPort)
	}

	if cfg.KafkaHost != "localhost:9092" {
		t.Fatalf("expected kafka host localhost:9092, got %q", cfg.KafkaHost)
	}
}

func TestLoadConfig_PanicWhenOrderServiceEmpty(t *testing.T) {
	unsetEnv(t, "ORDER_SERVICE_HOST", "KAFKA_HOST")

	mustSetenv(t, "KAFKA_HOST", "localhost")

	cfg := newTestConfig(t)

	mustPanic(t, func() {
		cfg.LoadConfig()
	})
}
