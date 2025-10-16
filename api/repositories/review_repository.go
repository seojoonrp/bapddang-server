// api/repositories/review_repository.go

package repositories

import (
	"context"

	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReviewRepository interface {
	SaveReview(review *models.Review) error
}

type reviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(coll *mongo.Collection) ReviewRepository {
	return &reviewRepository{collection: coll}
}

func (r *reviewRepository) SaveReview (review *models.Review) error {
	_, err := r.collection.InsertOne(context.TODO(), review)
	return err
}