// api/services/review_service.go

package services

import (
	"time"

	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewService interface {
	CreateReview(input models.CreateReviewInput, user models.User) (*models.Review, error)
}

type reviewService struct {
	reviewRepo repositories.ReviewRepository
}

func NewReviewService(repo repositories.ReviewRepository) ReviewService {
	return &reviewService{
		reviewRepo: repo,
	}
}

func (s *reviewService) CreateReview(input models.CreateReviewInput, user models.User) (*models.Review, error) {
	newReview := models.Review {
		ID: primitive.NewObjectID(),
		UserID: user.ID,
		Name: input.Name,
		Foods: input.Foods,
		Speed: input.Speed,
		MealTime: input.MealTime,
		Tags: input.Tags,
		ImageURL: input.ImageURL,
		Comment: input.Comment,
		Rating: input.Rating,
		Day: user.Day,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := s.reviewRepo.SaveReview(&newReview)
	if err != nil {
		return nil, err
	}
	return &newReview, err
}