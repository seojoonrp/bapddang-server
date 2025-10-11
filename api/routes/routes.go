package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(router *gin.Engine, db *mongo.Database) {
	apiV1 := router.Group("/api/v1")
	{
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/signup", handlers.SignUp)
			authRoutes.POST("/login", handlers.Login)
		}
	}
}