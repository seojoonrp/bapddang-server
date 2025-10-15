// routes/routes.go

// api 엔드포인트를 관리하는 라우터

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(router *gin.Engine, db *mongo.Database) {
	userCollection := db.Collection("users")
	handlers.SetUserCollection(userCollection)

	standardFoodCollection := db.Collection("standard-foods")
	handlers.SetStandardFoodCollection(standardFoodCollection)

	apiV1 := router.Group("/api/v1")
	{
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/signup", handlers.SignUp)
			authRoutes.POST("/login", handlers.Login)
		}

		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(userCollection))
		{
			protected.GET("/users/me/liked-foods", handlers.GetLikedFoodIDs)
			protected.POST("/foods/:foodId/like", handlers.LikeFood)
			protected.DELETE("/foods/:foodId/like", handlers.UnlikeFood)
		}

		apiV1.GET("/foods/:foodId", handlers.GetFoodByID)

		adminRoutes := apiV1.Group("/admin")
		{
			adminRoutes.POST("/new-food", handlers.CreateStandardFood)
		}
	}
}