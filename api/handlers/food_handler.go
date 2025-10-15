// api/handlers/food_handler.go

// 음식 관련 로직(음식 조회, 커스텀 음식 생성 등)을 처리하는 API 핸들러

package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

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

var customFoodCollection *mongo.Collection

func SetCustomFoodCollection (coll *mongo.Collection) {
	customFoodCollection = coll;
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

type NewCustomFoodInput struct {
	Name string `json:"name"`
}

func FindOrCreateCustomFood (ctx *gin.Context) {
	var input NewCustomFoodInput
	
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wrong input format"})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := userCtx.(models.User)

	var existingFood models.CustomFood
	filter := bson.M{"name": input.Name}
	err := customFoodCollection.FindOne(context.TODO(), filter).Decode(&existingFood)
	
	if err == mongo.ErrNoDocuments {
		newFood := models.CustomFood{
			ID: primitive.NewObjectID(),
			Name: input.Name,
			UsingUserIDs: []primitive.ObjectID{user.ID},
			CreatedAt: time.Now(),
		}

		_, err := customFoodCollection.InsertOne(context.TODO(), newFood)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom food"})
			return
		}

		ctx.JSON(http.StatusCreated, newFood)
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch custom food"})
		return
	}

	updateFilter := bson.M{"_id": existingFood.ID}
	update := bson.M{"$addToSet": bson.M{"usingUserIDs": user.ID}}

	_, err = customFoodCollection.UpdateOne(context.TODO(), updateFilter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update custom food users"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Custom food already exists, user added to usingUserIDs"})
}

type NewStandardFoodInput struct {
	Name string `json:"name"`
	ImageURL string `json:"imageURL"`
	Speed string `json:"speed"`
	Type string `json:"type"`
	Categories []string `json:"categories"`
}

// 어드민용 표준음식 추가 함수
func CreateStandardFood (ctx *gin.Context) {
	var input NewStandardFoodInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wrong input format"})
		return
	}

	newFood := models.StandardFood{
		ID: primitive.NewObjectID(),
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

	ctx.JSON(http.StatusCreated, newFood)
}