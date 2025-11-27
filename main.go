package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/routes"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/database"
)

func main() {
	config.LoadConfig()

	client, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(config.AppConfig.DBName)
	router := gin.Default()
	router.SetTrustedProxies(nil)
	routes.SetupRoutes(router, db)

	port := config.AppConfig.Port
	log.Println("Server started on port " + port + ".")
	router.Run(":" + port)
}