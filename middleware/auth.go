// middleware/auth.go
// JWT 인증 미들웨어

package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthMiddleware(userCollection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format not matched."})
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Token parsing error: %v", err) 
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userIDHex, ok := claims["sub"].(string)
			if (!ok) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				return
			}

			userID, err := primitive.ObjectIDFromHex(userIDHex)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			}

			var user models.User
			err = userCollection.FindOne(context.TODO(), primitive.M{"_id": userID}).Decode(&user)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}

			loc, err := time.LoadLocation("Asia/Seoul")
			if err != nil {
				ctx.Next()
				return
			}

			nowKST := time.Now().In(loc)
			createdAtKST := user.CreatedAt.In(loc)
			endDate := time.Date(nowKST.Year(), nowKST.Month(), nowKST.Day(), 0, 0, 0, 0, loc)
			startDate := time.Date(createdAtKST.Year(), createdAtKST.Month(), createdAtKST.Day(), 0, 0, 0, 0, loc)
			daysPassed := int(endDate.Sub(startDate).Hours() / 24)
			curDay := daysPassed + 1

			if user.Day < curDay {
				user.Day = curDay

				go func() {
					filter := bson.M{"_id": user.ID}
					update := bson.M{"$set": bson.M{"day": user.Day}}
					userCollection.UpdateOne(context.Background(), filter, update)
				}()
			}

			ctx.Set("currentUser", user)
			ctx.Next()
		} else {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}