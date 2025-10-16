// routes/routes.go

// api 엔드포인트를 관리하는 라우터

package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/handlers"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/middleware"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(router *gin.Engine, db *mongo.Database) {
	userCollection := db.Collection("users")
	userRepository := repositories.NewUserRepository(userCollection)
	userService := services.NewUserService(userRepository)
	userHandler := handlers.NewUserHandler(userService)

	standardFoodCollection := db.Collection("standard-foods")
	handlers.SetStandardFoodCollection(standardFoodCollection)

	customFoodCollection := db.Collection("custom-foods")
	handlers.SetCustomFoodCollection(customFoodCollection)

	reviewCollection := db.Collection("reviews")
	handlers.SetReviewCollection(reviewCollection)

	apiV1 := router.Group("/api/v1")
	{
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/signup", userHandler.SignUp)
			authRoutes.POST("/login", userHandler.Login)
		}

		protected := apiV1.Group("/")
		protected.Use(middleware.AuthMiddleware(userCollection))
		{
			protected.POST("/foods/:foodId/like", userHandler.LikeFood)
			protected.DELETE("/foods/:foodId/like", userHandler.UnlikeFood)
			protected.POST("/custom-foods", handlers.FindOrCreateCustomFood)
			protected.POST("/reviews", handlers.CreateReview)
		}

		apiV1.GET("/foods/:foodId", handlers.GetFoodByID)

		adminRoutes := apiV1.Group("/admin")
		{
			adminRoutes.POST("/new-food", handlers.CreateStandardFood)
		}
	}
}