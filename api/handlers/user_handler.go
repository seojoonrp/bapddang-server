// api/handlers/user_handler.go

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	var input models.SignUpInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.SignUp(input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var input models.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokenString, err := h.userService.Login(input)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *UserHandler) LikeFood(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	userID := userCtx.(models.User).ID

	foodIDHex := ctx.Param("foodID")
	foodID, _ := primitive.ObjectIDFromHex(foodIDHex)

	wasAdded, err := h.userService.LikeFood(userID, foodID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if wasAdded {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food liked successfully"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food is already liked"})
	}
}

func (h *UserHandler) UnlikeFood(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	userID := userCtx.(models.User).ID

	foodIDHex := ctx.Param("foodID")
	foodID, _ := primitive.ObjectIDFromHex(foodIDHex)

	wasRemoved, err := h.userService.UnlikeFood(userID, foodID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if wasRemoved {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food unliked successfully"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food was not liked"})
	}
}