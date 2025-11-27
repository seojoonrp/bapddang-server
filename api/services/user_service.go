// api/services/user_service.go

package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	SignUp(input models.SignUpInput) (*models.User, error)
	Login(input models.LoginInput) (string, error)
	LikeFood(userID, foodID primitive.ObjectID) (bool, error)
	UnlikeFood(userID, foodID primitive.ObjectID) (bool, error)
	GetLikedFoodIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error)
}

type userService struct {
	userRepo repositories.UserRepository
	foodRepo repositories.FoodRepository
}

func NewUserService(userRepo repositories.UserRepository, foodRepo repositories.FoodRepository) UserService {
	return &userService{userRepo: userRepo, foodRepo: foodRepo}
}

func (s *userService) SignUp(input models.SignUpInput) (*models.User, error) {
	_, err := s.userRepo.FindByEmail(input.Email)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:           primitive.NewObjectID(),
		UserName:     input.UserName,
		Email:        input.Email,
		Password:     string(hashedPassword),
		Day:          1,
		LikedFoodIDs: make([]primitive.ObjectID, 0),
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.Save(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) Login(input models.LoginInput) (string, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return "", err // Invalid email
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", err // Invalid password
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func (s *userService) LikeFood(userID, foodID primitive.ObjectID) (bool, error) {
	wasAdded, err := s.userRepo.AddLikedFood(userID, foodID)
	if err != nil {
		return false, err
	}

	return wasAdded, nil
}

func (s *userService) UnlikeFood(userID, foodID primitive.ObjectID) (bool, error) {
	wasRemoved, err := s.userRepo.RemoveLikedFood(userID, foodID)
	if err != nil {
		return false, err
	}

	return wasRemoved, nil
}

func (s *userService) GetLikedFoodIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	return s.userRepo.GetLikedFoodIDs(userID)
}