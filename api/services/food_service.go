// api/services/food_service.go

package services

import (
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodService interface {
	GetStandardFoodByID(id string) (*models.StandardFood, error)
	CreateStandardFood(input models.NewStandardFoodInput) (*models.StandardFood, error)
	FindOrCreateCustomFood(input models.NewCustomFoodInput, user models.User) (*models.CustomFood, error)
}

type foodService struct {
	foodRepo repositories.FoodRepository
}

func NewFoodService(repo repositories.FoodRepository) FoodService {
	return &foodService{
		foodRepo: repo,
	}
}

func (s *foodService) GetStandardFoodByID(id string) (*models.StandardFood, error) {
	foodID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return s.foodRepo.FindStandardFoodByID(foodID)
}

func (s *foodService) CreateStandardFood(input models.NewStandardFoodInput) (*models.StandardFood, error) {
	newFood := &models.StandardFood{
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
	err := s.foodRepo.SaveStandardFood(newFood)
	if err != nil {
		return nil, err
	}
	return newFood, nil
}

func (s *foodService) FindOrCreateCustomFood(input models.NewCustomFoodInput, user models.User) (*models.CustomFood, error) {
	existingFood, err := s.foodRepo.FindCustomFoodByName(input.Name)
	
	if err == mongo.ErrNoDocuments {
		newFood := &models.CustomFood{
			ID: primitive.NewObjectID(),
			Name: input.Name,
			UsingUserIDs: []primitive.ObjectID{user.ID},
			CreatedAt: time.Now(),
		}
		err := s.foodRepo.SaveCustomFood(newFood)
		if err != nil {
			return nil, err
		}
		return newFood, nil
	}

	if err != nil {
		return nil, err
	}

	err = s.foodRepo.AddUserToCustomFood(existingFood.ID, user.ID)
	if err != nil {
		return nil, err
	}

	return existingFood, nil
}