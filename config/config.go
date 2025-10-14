// config/config.go

// env 데이터 연결 및 설정을 관리

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	JWTSecret string
	Port string
	DBName string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using system variables.")
	}

	AppConfig = &Config{
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		JWTSecret: getEnv("JWT_SECRET_KEY", "default_secret"),
		Port: getEnv("PORT", "8080"),
		DBName: getEnv("DB_NAME", "bapddang-dev"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}