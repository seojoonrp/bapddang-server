package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.AppConfig.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}

	log.Println("Successfully connected to MongoDB.")

	db := client.Database(config.AppConfig.DBName)
	handlers.SetUserCollection(db.Collection("users"))

	r := gin.Default()

	r.POST("/signup", handlers.SignUp)
	r.POST("/login", handlers.Login)

	port := config.AppConfig.Port
	log.Println("Server started on port " + port + ".")
	r.Run(":" + port)
}