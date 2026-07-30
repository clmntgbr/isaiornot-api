package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL          string
	ClerkWebhookSecret   string
	Port                 string
	Environment          string
	ClerkSecretKey       string
	ClerkFrontendAPI     string
	RabbitMQURL          string
	RabbitMQSecretKey    string
	CORSAllowedOrigins   []string
	CORSAllowCredentials bool
	CORSAllowMethods     []string
	CORSAllowHeaders     []string
	CORSMaxAge           int
	RateLimitMax         int

	ExchangeName               string
	AnalyzeRequestQueueName    string
	MetadataAnalyzeQueueName   string
	MetadataDoneQueueName      string
	MetadataFailedQueueName    string
	HeuristicsAnalyzeQueueName string
	HeuristicsDoneQueueName    string
	HeuristicsFailedQueueName  string
	AiModelAnalyzeQueueName    string
	AiModelDoneQueueName       string
	AiModelFailedQueueName     string

	WorkerConcurrency int

	StorageEndpoint         string
	StorageInternalEndpoint string
	StorageRegion           string
	StorageAccessKey        string
	StorageSecretKey        string
	StorageBucket           string
	StorageThumbnailBucket  string
	StorageUsePathStyle     bool
	MinIOWebhookSecret      string

	SightengineAPIURL    string
	SightengineAPIUser   string
	SightengineAPISecret string

	CentrifugoURL         string
	CentrifugoAPIKey      string
	CentrifugoTokenSecret string
	CentrifugoPublicWSURL string

	StripeSecretKey     string
	StripeWebhookSecret string
	RedirectSuccessURL  string
	RedirectCancelURL   string
	RedirectPortalURL   string
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

		ExchangeName:               getEnv("EXCHANGE_NAME"),
		AnalyzeRequestQueueName:    getEnv("ANALYZE_REQUEST_QUEUE_NAME"),
		MetadataAnalyzeQueueName:   getEnv("METADATA_ANALYZE_QUEUE_NAME"),
		MetadataDoneQueueName:      getEnv("METADATA_DONE_QUEUE_NAME"),
		MetadataFailedQueueName:    getEnv("METADATA_FAILED_QUEUE_NAME"),
		HeuristicsAnalyzeQueueName: getEnv("HEURISTICS_ANALYZE_QUEUE_NAME"),
		HeuristicsDoneQueueName:    getEnv("HEURISTICS_DONE_QUEUE_NAME"),
		HeuristicsFailedQueueName:  getEnv("HEURISTICS_FAILED_QUEUE_NAME"),
		AiModelAnalyzeQueueName:    getEnv("AI_MODEL_ANALYZE_QUEUE_NAME"),
		AiModelDoneQueueName:       getEnv("AI_MODEL_DONE_QUEUE_NAME"),
		AiModelFailedQueueName:     getEnv("AI_MODEL_FAILED_QUEUE_NAME"),
		RabbitMQURL:                getEnv("RABBITMQ_URL"),
		RabbitMQSecretKey:          getEnvOrDefault("RABBITMQ_SECRET_KEY", ""),
		WorkerConcurrency:          getEnvIntOrDefault("WORKER_CONCURRENCY", 5),

		StorageEndpoint:         getEnvOrDefault("STORAGE_ENDPOINT", ""),
		StorageInternalEndpoint: getEnvOrDefault("STORAGE_INTERNAL_ENDPOINT", ""),
		StorageRegion:           getEnvOrDefault("STORAGE_REGION", "us-east-1"),
		StorageAccessKey:        getEnv("STORAGE_ACCESS_KEY"),
		StorageSecretKey:        getEnv("STORAGE_SECRET_KEY"),
		StorageBucket:           getEnv("STORAGE_BUCKET"),
		StorageThumbnailBucket:  getEnvOrDefault("STORAGE_THUMBNAIL_BUCKET", "thumbnails"),
		StorageUsePathStyle:     getEnvBool("STORAGE_USE_PATH_STYLE"),
		MinIOWebhookSecret:      getEnvOrDefault("MINIO_WEBHOOK_SECRET", ""),

		SightengineAPIURL:    getEnvOrDefault("SIGHTENGINE_API_URL", "https://api.sightengine.com/1.0/check.json"),
		SightengineAPIUser:   getEnvOrDefault("SIGHTENGINE_API_USER", ""),
		SightengineAPISecret: getEnvOrDefault("SIGHTENGINE_API_SECRET", ""),

		CentrifugoURL:         getEnvOrDefault("CENTRIFUGO_URL", "http://centrifugo:8000/api"),
		CentrifugoAPIKey:      getEnvOrDefault("CENTRIFUGO_API_KEY", ""),
		CentrifugoTokenSecret: getEnvOrDefault("CENTRIFUGO_TOKEN_SECRET", ""),
		CentrifugoPublicWSURL: getEnvOrDefault("CENTRIFUGO_PUBLIC_WS_URL", "ws://localhost:8000/connection/websocket"),

		StripeSecretKey:     getEnvOrDefault("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnvOrDefault("STRIPE_WEBHOOK_SECRET", ""),
		RedirectSuccessURL:  getEnvOrDefault("REDIRECT_SUCCESS_URL", "http://localhost:3000/subscription/success"),
		RedirectCancelURL:   getEnvOrDefault("REDIRECT_CANCEL_URL", "http://localhost:3000/subscription/failed"),
		RedirectPortalURL:   getEnvOrDefault("REDIRECT_PORTAL_URL", "http://localhost:3000/subscription"),
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
