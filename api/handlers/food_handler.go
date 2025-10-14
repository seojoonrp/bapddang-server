package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/mongo"
)

var standardFoodCollection *mongo.Collection

func SetStandardFoodCollection (coll *mongo.Collection) {
	standardFoodCollection = coll;
}

type NewFoodInput struct {
	Name string `json:"name"`
	ImageURL string `json:"imageURL"`
	Speed string `json:"speed"`
	Type string `json:"type"`
	Categories []string `json:"categories"`
}

func CreateStandardFood (ctx *gin.Context) {
	var input NewFoodInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wrong input format"})
		return
	}

	newFood := models.StandardFood{
		Name: input.Name,
		ImageURL: input.ImageURL,
		Speed: input.Speed,
		Type: input.Type,
		Categories: input.Categories,
		LikeCount: 0,
		ReviewCount: 0,
		AverageRating: 0.0,
		TrendScore: 0,
	}

	_, err := standardFoodCollection.InsertOne(context.TODO(), newFood)
	if err != nil {
		log.Printf("Error while creating new food: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new food"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "new food " + input.Name + " successfully created"})
}