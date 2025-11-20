// api/services/food_service.go

package services

import (
	"errors"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/models"
	"github.com/seojoonrp/bapddang-server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FoodService interface {
	GetStandardFoodByID(id string) (*models.StandardFood, error)
	CreateStandardFood(input models.NewStandardFoodInput) (*models.StandardFood, error)
	FindOrCreateCustomFood(input models.NewCustomFoodInput, user models.User) (*models.CustomFood, error)

	GetMainFeedFoods(foodType, speed string, foodCount int) ([]*models.StandardFood, error)
	ValidateFoods(names []string, userID primitive.ObjectID) ([]models.ValidationResult, error)
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
	_, err := s.foodRepo.FindStandardFoodByName(input.Name)
	if err == nil {
		return nil, errors.New("food already exists")
	}
	
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
	err = s.foodRepo.SaveStandardFood(newFood)
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

func (s *foodService) GetMainFeedFoods(foodType, speed string, foodCount int) ([]*models.StandardFood, error) {
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

	log.Printf("Selected %d foods: %+v", len(resultList), resultList)

	return resultList, nil
}

type matchCandidates struct {
	Score float64
	Output models.ValidationOutput
}
	
func (s *foodService) ValidateFoods(names []string, userID primitive.ObjectID) ([]models.ValidationResult, error) {
	results := make([]models.ValidationResult, 0, len(names))

	const similarityThreshold = 0.75
	const maxSuggestions = 3

	for _, name := range names {
		result := models.ValidationResult{OriginalName: name}

		standardFood, err := s.foodRepo.FindStandardFoodByName(name)
		if err == nil {
			result.Status = "ok"
			result.OkOutput = &models.ValidationOutput{
				ID: standardFood.ID,
				Name: standardFood.Name,
				Type: "standard",
			}
			results = append(results, result)
			continue
		}

		customFood, err := s.foodRepo.FindCustomFoodByName(name)
		if err == nil {
			result.Status = "ok"
			result.OkOutput = &models.ValidationOutput{
				ID: customFood.ID,
				Name: customFood.Name,
				Type: "custom",
			}
			results = append(results, result)
			continue
		}

		var candidates []matchCandidates

		s.cacheLock.RLock()

		for _, food := range s.standardFoodCache {
			score := utils.Score(name, food.Name)
			if score >= similarityThreshold {
				candidates = append(candidates, matchCandidates{
					Score: score,
					Output: models.ValidationOutput{
						ID: food.ID,
						Name: food.Name,
						Type: "standard",
					},
				})
			}
		}

		for _, food := range s.customFoodCache {
			score := utils.Score(name, food.Name)
			if score >= similarityThreshold {
				candidates = append(candidates, matchCandidates{
					Score: score,
					Output: models.ValidationOutput{
						ID: food.ID,
						Name: food.Name,
						Type: "custom",
					},
				})
			}
		}

		s.cacheLock.RUnlock()

		if len(candidates) == 0 {
			s.cacheLock.Lock()

			// TODO : Lock 걸면서 중복 생성됐는지 이중체크
			
			newCustomFood := &models.CustomFood{
				ID: primitive.NewObjectID(),
				Name: name,
				UsingUserIDs: []primitive.ObjectID{userID},
				CreatedAt: time.Now(),
			}

			err := s.foodRepo.SaveCustomFood(newCustomFood)
			if err != nil {
				s.cacheLock.Unlock()
				return nil, err
			}

			s.customFoodCache = append(s.customFoodCache, newCustomFood)

			result.Status = "new"
			result.NewOutput = &models.ValidationOutput{
				Name: name,
				Type: "new",
			}
			results = append(results, result)

			s.cacheLock.Unlock()
			continue
		}

		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})

		limit := min(len(candidates), maxSuggestions)
		for i := range limit {
			result.SuggestionOutputs = append(result.SuggestionOutputs, candidates[i].Output)
		}

		result.Status = "suggestion"
		results = append(results, result)
	}

	return results, nil
}

func (s *foodService) UpdateReviewStats(foodIDs []primitive.ObjectID, rating *int) error {
	return s.foodRepo.UpdateReviewStats(foodIDs, rating)
}