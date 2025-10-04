package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
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

func SignUp (c *gin.Context) {
	var input SignUpInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := models.User {
		Email: input.Email,
		Password: string(hashedPassword),
		UserName: input.UserName,
		CreatedAt: time.Now(),
		Day: 1,
	}

	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		log.Printf("Error while creating new user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

type LoginInput struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Login (c *gin.Context) {
	var input LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var user models.User
	err := userCollection.FindOne(context.TODO(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}