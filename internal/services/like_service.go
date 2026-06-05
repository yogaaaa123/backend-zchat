package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/repositories"
	"gorm.io/gorm"
)

type LikeService struct {
	likeRepo   *repositories.LikeRepository
	threadRepo *repositories.ThreadRepository
}

func NewLikeService(likeRepo *repositories.LikeRepository, threadRepo *repositories.ThreadRepository) *LikeService {
	return &LikeService{likeRepo: likeRepo, threadRepo: threadRepo}
}

type ToggleResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

func (s *LikeService) Toggle(userID, threadID uuid.UUID) (*ToggleResponse, error) {
	// Validate thread exists
	exists, err := s.threadRepo.FindThreadExists(threadID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrThreadNotFound
	}

	// Check if already liked
	existing, err := s.likeRepo.FindByUserAndThread(userID, threadID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if existing != nil {
		// Unlike
		if err := s.likeRepo.Delete(existing.ID); err != nil {
			return nil, err
		}
	} else {
		// Like
		like := &models.Like{
			ID:       uuid.New(),
			UserID:   userID,
			ThreadID: threadID,
		}
		if err := s.likeRepo.Create(like); err != nil {
			return nil, err
		}
	}

	// Get updated count
	count, err := s.likeRepo.CountByThread(threadID)
	if err != nil {
		return nil, err
	}

	return &ToggleResponse{
		Liked:     existing == nil,
		LikeCount: int(count),
	}, nil
}
