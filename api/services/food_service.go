// api/services/food_service.go

package services

import (
	"log"
	"math/rand"
	"sync"
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

	GetMainFeedFoods(foodType, speed string) ([]*models.StandardFood, error)
	ValidateFoods(names []string) ([]models.ValidationResult, error)
	UpdateReviewStats(foodIDs []primitive.ObjectID, rating *int) error
}

type foodService struct {
	foodRepo repositories.FoodRepository
	standardFoodCache []*models.StandardFood
	customFoodCache []*models.CustomFood
	cacheLock sync.RWMutex
}

func NewFoodService(foodRepo repositories.FoodRepository) FoodService {
	allStandardFoods, err := foodRepo.GetAllStandardFoods()
	if err != nil {
		log.Fatal("FATAL: Failed to load standard food cache: ", err)
	}
	allCustomFoods, err := foodRepo.GetAllCustomFoods()
	if err != nil {
		log.Fatal("FATAL: Failed to load custom food cache: ", err)
	}

	return &foodService{
		foodRepo: foodRepo,
		standardFoodCache: allStandardFoods,
		customFoodCache: allCustomFoods,
		cacheLock: sync.RWMutex{},
	}
}

func (s *foodService) GetStandardFoodByID(id string) (*models.StandardFood, error) {
	foodID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	food, _ := s.foodRepo.FindStandardFoodByID(foodID)

	if food.RatedReviewCount > 0 {
		food.AverageRating = float64(food.TotalRating) / float64(food.RatedReviewCount)
	} else {
		food.AverageRating = 0.0
	}

	return food, nil
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
		RatedReviewCount: 0,
		TotalRating: 0,
		AverageRating: 0.0,
		TrendScore: 0,
	}
	err := s.foodRepo.SaveStandardFood(newFood)
	if err != nil {
		return nil, err
	}

	s.cacheLock.Lock()
	s.standardFoodCache = append(s.standardFoodCache, newFood)
	s.cacheLock.Unlock()
	
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

		s.cacheLock.Lock()
		s.customFoodCache = append(s.customFoodCache, newFood)
		s.cacheLock.Unlock()

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

func (s *foodService) GetMainFeedFoods(foodType, speed string) ([]*models.StandardFood, error) {
	foodCount := 10

	s.cacheLock.RLock()
	defer s.cacheLock.RUnlock()
	
	candidates := make([]*models.StandardFood, 0)
	for _, food := range s.standardFoodCache {
		if food.Type == foodType && food.Speed == speed {
			candidates = append(candidates, food)
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	resultList := make([]*models.StandardFood, 0, foodCount)
	usedCategories := make(map[string]bool)

	for _, food := range candidates {
		if len(resultList) >= foodCount {
			break
		}
		if len(food.Categories) == 0 {
			resultList = append(resultList, food)
			continue
		}

		uniqueCategory := true
		for _, category := range food.Categories {
			if usedCategories[category] {
				uniqueCategory = false
				break
			}
		}
		if uniqueCategory {
			resultList = append(resultList, food)
			for _, category := range food.Categories {
				usedCategories[category] = true
			}
		}
	}

	for _, food := range resultList {
		if food.RatedReviewCount > 0 {
			food.AverageRating = float64(food.TotalRating) / float64(food.RatedReviewCount)
		} else {
			food.AverageRating = 0.0
		}
	}

	return resultList, nil
}

func (s *foodService) ValidateFoods(names []string) ([]models.ValidationResult, error) {
	results := make([]models.ValidationResult, 0, len(names))

	for _, name := range names {
		result := models.ValidationResult{OriginalName: name}

		standardFood, err := s.foodRepo.FindStandardFoodByName(name)
		if err == nil {
			result.Status = "ok"
			result.Food = standardFood
			results = append(results, result)
			continue
		}

		customFood, err := s.foodRepo.FindCustomFoodByName(name)
		if err == nil {
			result.Status = "ok"
			result.Food = customFood
			results = append(results, result)
			continue
		}

		standardSuggestion, err := s.foodRepo.SearchSimilarStandardFood(name)
		if err == nil {
			result.Status = "suggestion"
			result.Suggestion = standardSuggestion
			results = append(results, result)
			continue
		}

		customSuggestion, err := s.foodRepo.SearchSimilarCustomFood(name)
		if err == nil {
			result.Status = "suggestion"
			result.Suggestion = customSuggestion
			results = append(results, result)
			continue
		}

		result.Status = "new"
		results = append(results, result)
	}

	return results, nil
}

func (s *foodService) UpdateReviewStats(foodIDs []primitive.ObjectID, rating *int) error {
	return s.foodRepo.UpdateReviewStats(foodIDs, rating)
}