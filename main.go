package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	userCollection := client.Database("bapddang-dev").Collection("users")
	handlers.SetUserCollection(userCollection)

	r := gin.Default()

	r.POST("/signup", handlers.SignUp)
	r.POST("/login", handlers.Login)

	// protected := r.Group("/api")
	// protected.Use(middleware.AuthMiddleware(userCollection)) {
		
	// }

	log.Println("Server started on port 8080.")
	r.Run(":8080")
}