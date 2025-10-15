// api/handlers/user_handler.go

// 유저 관련 로직(로그인, 회원가입, 데이터 fetching)을 처리하는 API 핸들러

package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection

func SetUserCollection (coll *mongo.Collection) {
	userCollection = coll;
}

type SignUpInput struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	UserName string `json:"userName" binding:"required"`
}

func SignUp (ctx *gin.Context) {
	var input SignUpInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := models.User {
		ID: primitive.NewObjectID(),
		UserName: input.UserName,
		Email: input.Email,
		Password: string(hashedPassword),
		Day: 1,
		LikedFoodIDs: make([]primitive.ObjectID, 0),
		CreatedAt: time.Now(),
	}

	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		log.Printf("Error while creating new user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new user"})
		return
	}

	ctx.JSON(http.StatusCreated, newUser)
}

type LoginInput struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login (ctx *gin.Context) {
	var input LoginInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}
	
	// 로그인시 JWT토큰을 발급
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func GetLikedFoodIDs(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(models.User)
	ctx.JSON(http.StatusOK, user.LikedFoodIDs)
}

func LikeFood(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	userID := user.ID

	foodIDStr := ctx.Param("foodId")
	foodID, err := primitive.ObjectIDFromHex(foodIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID"})
		return
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$addToSet": bson.M{"likedFoodIDs": foodID}}

	result, err := userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user likes"})
		return
	}

	if result.ModifiedCount == 0 {
		ctx.JSON(http.StatusOK, gin.H{"message": "Food was already liked"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully liked the food"})
}

func UnlikeFood(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	userID := user.ID

	foodIDStr := ctx.Param("foodId")
	foodID, err := primitive.ObjectIDFromHex(foodIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid food ID format"})
		return
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$pull": bson.M{"likedFoodIDs": foodID}}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user likes"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully unliked the food"})
}