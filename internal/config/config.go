// Package config provides centralized configuration management for ClawTeam
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	API     APIConfig
	LLM     LLMConfig
	DB      DBConfig
	NATS    NATSConfig
	Qdrant  QdrantConfig
	Execution ExecutionConfig
	Logging LoggingConfig
	Feature FeatureConfig
}

// APIConfig holds API server configuration
type APIConfig struct {
	Host    string
	Port    int
	Workers int
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	AnthropicAPIKey string
	OpenAIAPIKey    string
	DefaultModel    string
}

// DBConfig holds database configuration
type DBConfig struct {
	URL         string
	PoolSize    int
	MaxOverflow int
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL            string
	User           string
	Password       string
	SubjectPrefix  string
}

// QdrantConfig holds Qdrant vector database configuration
type QdrantConfig struct {
	URL     string
	APIKey  string
}

// ExecutionConfig holds execution settings
type ExecutionConfig struct {
	JobTimeoutSeconds int
	MaxConcurrentJobs int
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string // json, text
}

// FeatureConfig holds feature flags
type FeatureConfig struct {
	EnableMetrics bool
	EnableTracing bool
}

// defaultConfig returns configuration with default values
func defaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Host:    "0.0.0.0",
			Port:    8000,
			Workers: 1,
		},
		LLM: LLMConfig{
			DefaultModel: "claude-3-5-sonnet-20241022",
		},
		DB: DBConfig{
			URL:         "postgresql://clawteam:password@localhost:5432/clawteam",
			PoolSize:    10,
			MaxOverflow: 20,
		},
		NATS: NATSConfig{
			URL:           "nats://localhost:4222",
			SubjectPrefix:  "clawteam",
		},
		Qdrant: QdrantConfig{
			URL: "http://localhost:6333",
		},
		Execution: ExecutionConfig{
			JobTimeoutSeconds: 60,
			MaxConcurrentJobs: 10,
		},
		Logging: LoggingConfig{
			Level:  "INFO",
			Format: "json",
		},
		Feature: FeatureConfig{
			EnableMetrics: true,
			EnableTracing: false,
		},
	}
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := defaultConfig()

	// API Configuration
	if v := getEnv("CLAWTEAM_API_HOST"); v != "" {
		cfg.API.Host = v
	}
	if v := getEnv("CLAWTEAM_API_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAWTEAM_API_PORT: %w", err)
		}
		cfg.API.Port = port
	}
	if v := getEnv("CLAWTEAM_API_WORKERS"); v != "" {
		workers, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAWTEAM_API_WORKERS: %w", err)
		}
		cfg.API.Workers = workers
	}

	// LLM Configuration
	cfg.LLM.AnthropicAPIKey = getEnv("CLAWTEAM_ANTHROPIC_API_KEY")
	cfg.LLM.OpenAIAPIKey = getEnv("CLAWTEAM_OPENAI_API_KEY")
	if v := getEnv("CLAWTEAM_DEFAULT_MODEL"); v != "" {
		cfg.LLM.DefaultModel = v
	}

	// Database Configuration
	if v := getEnv("CLAWTEAM_DATABASE_URL"); v != "" {
		cfg.DB.URL = v
	}
	if v := getEnv("CLAWTEAM_DATABASE_POOL_SIZE"); v != "" {
		poolSize, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAWTEAM_DATABASE_POOL_SIZE: %w", err)
		}
		cfg.DB.PoolSize = poolSize
	}
	if v := getEnv("CLAWTEAM_DATABASE_MAX_OVERFLOW"); v != "" {
		maxOverflow, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAWTEAM_DATABASE_MAX_OVERFLOW: %w", err)
		}
		cfg.DB.MaxOverflow = maxOverflow
	}

	// NATS Configuration
	if v := getEnv("CLAWTEAM_NATS_URL"); v != "" {
		cfg.NATS.URL = v
	}
	cfg.NATS.User = getEnv("CLAWTEAM_NATS_USER")
	cfg.NATS.Password = getEnv("CLAWTEAM_NATS_PASSWORD")
	if v := getEnv("CLAWTEAM_NATS_SUBJECT_PREFIX"); v != "" {
		cfg.NATS.SubjectPrefix = v
	}

	// Qdrant Configuration
	if v := getEnv("CLAWTEAM_QDRANT_URL"); v != "" {
		cfg.Qdrant.URL = v
	}
	cfg.Qdrant.APIKey = getEnv("CLAWTEAM_QDRANT_API_KEY")

	// Execution Configuration
	if v := getEnv("CLAWTEAM_JOB_TIMEOUT_SECONDS"); v != "" {
		timeout, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAWTEAM_JOB_TIMEOUT_SECONDS: %w", err)
		}
		cfg.Execution.JobTimeoutSeconds = timeout
	}
	if v := getEnv("CLAWTEAM_MAX_CONCURRENT_JOBS"); v != "" {
		maxJobs, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid CLAWTEAM_MAX_CONCURRENT_JOBS: %w", err)
		}
		cfg.Execution.MaxConcurrentJobs = maxJobs
	}

	// Logging Configuration
	if v := getEnv("CLAWTEAM_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = strings.ToUpper(v)
	}
	if v := getEnv("CLAWTEAM_LOG_FORMAT"); v != "" {
		cfg.Logging.Format = strings.ToLower(v)
	}

	// Feature Flags
	if v := getEnv("CLAWTEAM_ENABLE_METRICS"); v != "" {
		cfg.Feature.EnableMetrics = strings.ToLower(v) == "true"
	}
	if v := getEnv("CLAWTEAM_ENABLE_TRACING"); v != "" {
		cfg.Feature.EnableTracing = strings.ToLower(v) == "true"
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a file (supports .env, .json, .yaml)
func LoadFromFile(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	// Set defaults
	v.SetDefault("api.host", "0.0.0.0")
	v.SetDefault("api.port", 8000)
	v.SetDefault("api.workers", 1)
	v.SetDefault("llm.default_model", "claude-3-5-sonnet-20241022")
	v.SetDefault("database.pool_size", 10)
	v.SetDefault("database.max_overflow", 20)
	v.SetDefault("execution.job_timeout_seconds", 60)
	v.SetDefault("execution.max_concurrent_jobs", 10)
	v.SetDefault("log.level", "INFO")
	v.SetDefault("log.format", "json")
	v.SetDefault("feature.enable_metrics", true)
	v.SetDefault("feature.enable_tracing", false)

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("CLAWTEAM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// getEnv retrieves an environment variable or returns empty string
func getEnv(key string) string {
	return os.Getenv(key)
}

// MustLoad loads configuration or panics
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Check required API keys for LLM
	if c.LLM.AnthropicAPIKey == "" && c.LLM.OpenAIAPIKey == "" {
		return fmt.Errorf("at least one LLM API key (CLAWTEAM_ANTHROPIC_API_KEY or CLAWTEAM_OPENAI_API_KEY) is required")
	}

	// Validate log level
	validLevels := map[string]bool{
		"TRACE": true,
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
		"FATAL": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	// Validate log format
	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", c.Logging.Format)
	}

	return nil
}

// GetAPIAddr returns the API address in host:port format
func (c *Config) GetAPIAddr() string {
	return fmt.Sprintf("%s:%d", c.API.Host, c.API.Port)
}
