package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	ClerkWebhookSecret   string
	Port                 string
	Environment          string
	ClerkSecretKey       string
	ClerkFrontendAPI     string
	CORSAllowedOrigins   []string
	CORSAllowCredentials bool
	CORSAllowMethods     []string
	CORSAllowHeaders     []string
	CORSMaxAge           int
	RateLimitMax         int
	RabbitMQURL          string
	RabbitMQExchange     string
	RabbitMQQueue        string
	RabbitMQRoutingKey   string
	RabbitMQRetryTTLMS   int
	WorkerMaxRetries     int
	OutboxPollInterval   time.Duration
	WorkerConcurrency    int

	MetadataAnalyzeQueue      string
	MetadataAnalyzeRoutingKey string

	HeuristicsAnalyzeQueue      string
	HeuristicsAnalyzeRoutingKey string

	AIModelAnalyzeQueue      string
	AIModelAnalyzeRoutingKey string

	SightengineAPIURL    string
	SightengineAPIUser   string
	SightengineAPISecret string

	StorageEndpoint         string
	StorageInternalEndpoint string
	StorageRegion           string
	StorageAccessKey        string
	StorageSecretKey        string
	StorageBucket           string
	StorageThumbnailBucket  string
	StorageUsePathStyle     bool
	MinIOWebhookSecret      string

	CentrifugoURL         string
	CentrifugoAPIKey      string
	CentrifugoTokenSecret string
	CentrifugoPublicWSURL string

	StripeSecretKey      string
	StripeWebhookSecret  string
	RedirectSuccessURL   string
	RedirectCancelURL    string
	RedirectPortalURL    string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL:          getEnv("DATABASE_URL"),
		ClerkWebhookSecret:   getEnv("CLERK_WEBHOOK_SECRET"),
		Port:                 getEnv("PORT"),
		Environment:          getEnv("GO_ENV"),
		ClerkSecretKey:       getEnv("CLERK_SECRET_KEY"),
		ClerkFrontendAPI:     getEnv("CLERK_FRONTEND_API"),
		CORSAllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS"), ","),
		CORSAllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS"),
		CORSAllowMethods:     strings.Split(getEnv("CORS_ALLOW_METHODS"), ","),
		CORSAllowHeaders:     strings.Split(getEnv("CORS_ALLOW_HEADERS"), ","),
		CORSMaxAge:           getEnvInt("CORS_MAX_AGE"),
		RateLimitMax:         getEnvInt("RATE_LIMIT_MAX"),
		RabbitMQURL:          getEnv("RABBITMQ_URL"),
		RabbitMQExchange:     getEnvOrDefault("RABBITMQ_EXCHANGE", "domain.events"),
		RabbitMQQueue:        getEnvOrDefault("RABBITMQ_QUEUE", "domain.events"),
		RabbitMQRoutingKey: getEnvOrDefault(
			"RABBITMQ_ROUTING_KEY",
			"user.#,media.*.v1,scan.#,media.analyze.metadata.done.v1,media.analyze.heuristics.done.v1,media.analyze.ai_model.done.v1",
		),
		RabbitMQRetryTTLMS: getEnvIntOrDefault("RABBITMQ_RETRY_TTL_MS", 30000),
		WorkerMaxRetries:   getEnvIntOrDefault("WORKER_MAX_RETRIES", 3),
		OutboxPollInterval: getEnvDuration("OUTBOX_POLL_INTERVAL", 2*time.Second),
		WorkerConcurrency:  getEnvIntOrDefault("WORKER_CONCURRENCY", 4),

		MetadataAnalyzeQueue:      getEnvOrDefault("METADATA_ANALYZE_QUEUE", "metadata.analyze"),
		MetadataAnalyzeRoutingKey: getEnvOrDefault("METADATA_ANALYZE_ROUTING_KEY", "media.analyze.metadata.v1"),

		HeuristicsAnalyzeQueue:      getEnvOrDefault("HEURISTICS_ANALYZE_QUEUE", "heuristics.analyze"),
		HeuristicsAnalyzeRoutingKey: getEnvOrDefault("HEURISTICS_ANALYZE_ROUTING_KEY", "media.analyze.heuristics.v1"),

		AIModelAnalyzeQueue:      getEnvOrDefault("AI_MODEL_ANALYZE_QUEUE", "ai_model.analyze"),
		AIModelAnalyzeRoutingKey: getEnvOrDefault("AI_MODEL_ANALYZE_ROUTING_KEY", "media.analyze.ai_model.v1"),

		SightengineAPIURL:    getEnvOrDefault("SIGHTENGINE_API_URL", "https://api.sightengine.com/1.0/check.json"),
		SightengineAPIUser:   getEnvOrDefault("SIGHTENGINE_API_USER", ""),
		SightengineAPISecret: getEnvOrDefault("SIGHTENGINE_API_SECRET", ""),

		StorageEndpoint:         getEnv("STORAGE_ENDPOINT"),
		StorageInternalEndpoint: getEnvOrDefault("STORAGE_INTERNAL_ENDPOINT", ""),
		StorageRegion:           getEnvOrDefault("STORAGE_REGION", "us-east-1"),
		StorageAccessKey:        getEnv("STORAGE_ACCESS_KEY"),
		StorageSecretKey:        getEnv("STORAGE_SECRET_KEY"),
		StorageBucket:           getEnv("STORAGE_BUCKET"),
		StorageThumbnailBucket:  getEnvOrDefault("STORAGE_THUMBNAIL_BUCKET", "thumbnails"),
		StorageUsePathStyle:     getEnvBoolOrDefault("STORAGE_USE_PATH_STYLE", true),
		MinIOWebhookSecret:      getEnvOrDefault("MINIO_WEBHOOK_SECRET", ""),

		CentrifugoURL:         getEnv("CENTRIFUGO_URL"),
		CentrifugoAPIKey:      getEnv("CENTRIFUGO_API_KEY"),
		CentrifugoTokenSecret: getEnv("CENTRIFUGO_TOKEN_SECRET"),
		CentrifugoPublicWSURL: getEnvOrDefault("CENTRIFUGO_PUBLIC_WS_URL", ""),

		StripeSecretKey:     getEnvOrDefault("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnvOrDefault("STRIPE_WEBHOOK_SECRET", ""),
		RedirectSuccessURL:  getEnvOrDefault("REDIRECT_SUCCESS_URL", "http://localhost:3001/subscription/success"),
		RedirectCancelURL:   getEnvOrDefault("REDIRECT_CANCEL_URL", "http://localhost:3001/subscription/failed"),
		RedirectPortalURL:   getEnvOrDefault("REDIRECT_PORTAL_URL", "http://localhost:3001/subscription"),
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	log.Panicf("required environment variable %s is not set", key)
	return ""
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		return false
	}

	return value == "true"
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true"
}

func getEnvInt(key string) int {
	value := os.Getenv(key)
	if value == "" {
		log.Panicf("required environment variable %s is not set", key)
		return 0
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid integer for %s: %q", key, value)
		return 0
	}

	return parsedValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		log.Panicf("invalid integer for %s: %q", key, value)
		return 0
	}
	return parsedValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Panicf("invalid duration for %s: %q", key, value)
		return 0
	}
	return parsed
}
