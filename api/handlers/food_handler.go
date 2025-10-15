package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var standardFoodCollection *mongo.Collection

func SetStandardFoodCollection (coll *mongo.Collection) {
	standardFoodCollection = coll;
}

func GetFoodByID(ctx *gin.Context) {
	foodIDStr := ctx.Param("foodId")
	foodID, err := primitive.ObjectIDFromHex(foodIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID format"})
		return
	}

	var food models.StandardFood
	filter := bson.M{"_id": foodID}
	
	err = standardFoodCollection.FindOne(context.TODO(), filter).Decode(&food)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Food not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch food data"})
		return
	}

	ctx.JSON(http.StatusOK, food)
}


type NewFoodInput struct {
	Name string `json:"name"`
	ImageURL string `json:"imageURL"`
	Speed string `json:"speed"`
	Type string `json:"type"`
	Categories []string `json:"categories"`
}

// 어드민용 표준음식 추가 함수
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