package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var reviewCollection *mongo.Collection

func SetReviewCollection (coll *mongo.Collection) {
	reviewCollection = coll;
}

type CreateReviewInput struct {
	Name string `json:"name" binding:"required"`
	Foods []models.ReviewedFoodItem `json:"foods" binding:"required"`
	Speed string `json:"speed" binding:"required"`
	MealTime string `json:"mealTime" binding:"required"`
	Tags []string `json:"tags"`
	ImageURL string `json:"imageUrl"`
	Comment string `json:"comment"`
	Rating int `json:"rating" binding:"required"`
}

func CreateReview (ctx *gin.Context) {
	var input CreateReviewInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userCtx.(models.User)

	newReview := models.Review {
		ID: primitive.NewObjectID(),
		UserID: user.ID,
		Name: input.Name,
		Foods: input.Foods,
		Speed: input.Speed,
		MealTime: input.MealTime,
		Tags: input.Tags,
		ImageURL: input.ImageURL,
		Comment: input.Comment,
		Rating: input.Rating,
		Day: user.Day,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := reviewCollection.InsertOne(context.TODO(), newReview)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	ctx.JSON(http.StatusCreated, newReview)
}