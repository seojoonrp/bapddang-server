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

	standardFoodCollection := db.Collection("standard-foods")
	customFoodCollection := db.Collection("custom-foods")
	foodRepository := repositories.NewFoodRepository(standardFoodCollection, customFoodCollection)

	reviewCollection := db.Collection("reviews")
	reviewRepository := repositories.NewReviewRepository(reviewCollection)

	userService := services.NewUserService(userRepository, foodRepository)
	userHandler := handlers.NewUserHandler(userService)

	foodService := services.NewFoodService(foodRepository)
	foodHandler := handlers.NewFoodHandler(foodService)

	reviewService := services.NewReviewService(reviewRepository, foodRepository)
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
			protected.GET("/auth/me", userHandler.GetMe)
			
			protected.GET("/liked-foods", userHandler.GetLikedFoods)

			protected.POST("/foods/:foodID/like", userHandler.LikeFood)
			protected.DELETE("/foods/:foodID/like", userHandler.UnlikeFood)
			protected.POST("/custom-foods", foodHandler.FindOrCreateCustomFood)
			protected.POST("/foods/validate", foodHandler.ValidateFoods)
			
			protected.POST("/reviews", reviewHandler.CreateReview)
			protected.GET("/reviews/me", reviewHandler.GetMyReviewsByDay)
		}

		apiV1.GET("/foods/:foodID", foodHandler.GetStandardFoodByID)
		apiV1.GET("/foods/main-feed", foodHandler.GetMainFeedFoods)

		adminRoutes := apiV1.Group("/admin")
		{
			adminRoutes.POST("/new-food", foodHandler.CreateStandardFood)
		}
	}
}