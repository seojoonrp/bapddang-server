// config/config.go

// env 데이터 연결 및 설정을 관리

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	MongoURI string
	DBName string

	JWTSecret string

	GoogleWebClientID string

	AWSAccessKeyID string
	AWSSecretAccessKey string
	AWSS3BucketName string
	AWSRegion string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using system variables.")
	}

	AppConfig = &Config{
		Port: getEnv("PORT", "8080"),

		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName: getEnv("DB_NAME", "bapddang-dev"),
		
		JWTSecret: getEnv("JWT_SECRET_KEY", "default_secret"),

		GoogleWebClientID: getEnv("GOOGLE_WEB_CLIENT_ID", ""),

		AWSAccessKeyID: getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
		AWSS3BucketName: getEnv("AWS_S3_BUCKET_NAME", ""),
		AWSRegion: getEnv("AWS_REGION", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}