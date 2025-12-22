// models/user.go

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	LoginMethodEmail = "email"
	LoginMethodGoogle = "google"
	LoginMethodKakao = "kakao"
	LoginMethodApple = "apple"
)

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserName string `bson:"userName" json:"userName" binding:"required"`
	Email string `bson:"email" json:"email" binding:"required"`
	Password string `bson:"password" json:"-"`
	LoginMethod string `bson:"loginMethod" json:"loginMethod"`
	Day int `bson:"day" json:"day"`
	LikedFoodIDs []primitive.ObjectID `bson:"likedFoodIDs" json:"likedFoodIDs"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type SignUpInput struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	UserName string `json:"userName" binding:"required"`
}

type LoginInput struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}