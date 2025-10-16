// api/handlers/food_handler.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodHandler struct {
	foodService services.FoodService
}

func NewFoodHandler(foodService services.FoodService) *FoodHandler {
	return &FoodHandler{
		foodService: foodService,
	}
}

func (h *FoodHandler) GetStandardFoodByID(ctx *gin.Context) {
	foodIDStr := ctx.Param("foodId")

	food, err := h.foodService.GetStandardFoodByID(foodIDStr)
	if err != nil {
		// 에러 추상화하기 귀찮다
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID"})
		return
	}

	ctx.JSON(http.StatusOK, food)
}

func (h *FoodHandler) CreateStandardFood(ctx *gin.Context) {
	var input models.NewStandardFoodInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	newFood, err := h.foodService.CreateStandardFood(input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create food"})
		return
	}

	ctx.JSON(http.StatusCreated, newFood)
}

func (h *FoodHandler) FindOrCreateCustomFood(ctx *gin.Context) {
	var input models.NewCustomFoodInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	customFood, err := h.foodService.FindOrCreateCustomFood(input, user.(models.User))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create custom food"})
		return
	}

	ctx.JSON(http.StatusOK, customFood)
}