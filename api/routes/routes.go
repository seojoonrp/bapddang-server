// routes/routes.go

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
	customFoodCollection := db.Collection("custom-foods")
	foodRepository := repositories.NewFoodRepository(standardFoodCollection, customFoodCollection)
	foodService := services.NewFoodService(foodRepository)
	foodHandler := handlers.NewFoodHandler(foodService)

	reviewCollection := db.Collection("reviews")
	reviewRepository := repositories.NewReviewRepository(reviewCollection)
	reviewService := services.NewReviewService(reviewRepository)
	reviewHandler := handlers.NewReviewHandler(reviewService)

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
			protected.POST("/custom-foods", foodHandler.FindOrCreateCustomFood)
			protected.POST("/foods/validate", foodHandler.ValidateFoods)
			protected.POST("/reviews", reviewHandler.CreateReview)
			protected.GET("/reviews/me", reviewHandler.GetMyReviewsByDay)
		}

		apiV1.GET("/foods/:foodId", foodHandler.GetStandardFoodByID)

		adminRoutes := apiV1.Group("/admin")
		{
			adminRoutes.POST("/new-food", foodHandler.CreateStandardFood)
		}
	}
}