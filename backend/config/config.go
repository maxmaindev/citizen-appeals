package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database       DatabaseConfig
	Server         ServerConfig
	JWT            JWTConfig
	CORS           CORSConfig
	Upload         UploadConfig
	AWS            AWSConfig
	Redis          RedisConfig
	Classification ClassificationConfig
	MongoDB        MongoDBConfig
	Env            string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Host string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type UploadConfig struct {
	MaxSize    int64
	UploadPath string
}

type AWSConfig struct {
	Region          string
	BucketName      string
	AccessKeyID     string
	SecretAccessKey string
	UseS3           bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ClassificationConfig struct {
	ServiceURL string
	Enabled    bool
}

type MongoDBConfig struct {
	URI        string
	Database   string
	AuthSource string
	Enabled    bool
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	jwtExpiration, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	maxUploadSize, err := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "5242880"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_UPLOAD_SIZE: %w", err)
	}

	useS3, _ := strconv.ParseBool(getEnv("USE_S3", "false"))

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "citizen_appeals"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			Expiration: jwtExpiration,
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173"}),
		},
		Upload: UploadConfig{
			MaxSize:    maxUploadSize,
			UploadPath: getEnv("UPLOAD_PATH", "./uploads"),
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			BucketName:      getEnv("AWS_BUCKET_NAME", ""),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
			UseS3:           useS3,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Classification: ClassificationConfig{
			ServiceURL: getEnv("CLASSIFICATION_SERVICE_URL", "http://localhost:8000"),
			Enabled:    getEnv("CLASSIFICATION_ENABLED", "true") == "true",
		},
		MongoDB: MongoDBConfig{
			URI:        getEnv("MONGODB_URI", "mongodb://admin:admin@localhost:27017"),
			Database:   getEnv("MONGODB_DATABASE", "citizen_appeals_ml"),
			AuthSource: getEnv("MONGODB_AUTH_SOURCE", "admin"),
			Enabled:    getEnv("MONGODB_ENABLED", "true") == "true",
		},
		Env: getEnv("ENV", "development"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Support comma-separated values
		values := strings.Split(value, ",")
		result := make([]string, 0, len(values))
		for _, v := range values {
			v = strings.TrimSpace(v)
			if v != "" {
				result = append(result, v)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
