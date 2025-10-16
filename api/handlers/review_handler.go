// api/handlers/review_handler.go

// 리뷰 관련 로직(리뷰 생성, 수정, 삭제, 조회 등)을 처리하는 API 핸들러

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seojoonrp/bapddang-server/api/services"
	"github.com/seojoonrp/bapddang-server/models"
)

type ReviewHandler struct {
	reviewService services.ReviewService
}

func NewReviewHandler(reviewService services.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		reviewService: reviewService,
	}
}

func (h *ReviewHandler) CreateReview (ctx *gin.Context) {
	var input models.CreateReviewInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)

	newReview, err := h.reviewService.CreateReview(input, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create review"})
		return
	}

	ctx.JSON(http.StatusCreated, newReview)
}