package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"github.com/satria/obrolan-api/internal/repositories"
	"github.com/satria/obrolan-api/internal/utils"
	"gorm.io/gorm"
)

var (
	ErrThreadNotFound = errors.New("thread not found")
	ErrForbidden      = errors.New("forbidden")
)

type ThreadService struct {
	threadRepo    *repositories.ThreadRepository
	likeRepo      *repositories.LikeRepository
	imageService  ImageService
}

func NewThreadService(threadRepo *repositories.ThreadRepository, likeRepo *repositories.LikeRepository, imageService ImageService) *ThreadService {
	return &ThreadService{threadRepo: threadRepo, likeRepo: likeRepo, imageService: imageService}
}

type CreateThreadRequest struct {
	Title    string `json:"title" binding:"required,min=5,max=255"`
	Content  string `json:"content" binding:"required,min=10"`
	ImageURL string `json:"image_url"`
}

type UpdateThreadRequest struct {
	Title    string `json:"title" binding:"required,min=5,max=255"`
	Content  string `json:"content" binding:"required,min=10"`
	ImageURL string `json:"image_url"`
}

type ThreadResponse struct {
	ID           uuid.UUID          `json:"id"`
	UserID       uuid.UUID          `json:"user_id"`
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	ImageURL     string             `json:"image_url,omitempty"`
	IsLiked      bool               `json:"is_liked"`
	LikeCount    int                `json:"like_count"`
	CommentCount int                `json:"comment_count"`
	User         *UserBriefResponse `json:"user"`
	CreatedAt    string             `json:"created_at"`
	UpdatedAt    string             `json:"updated_at"`
}

type UserBriefResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

func modelToThreadResponse(thread *models.Thread, isLiked bool) *ThreadResponse {
	userBrief := &UserBriefResponse{
		ID:        thread.User.ID,
		Username:  thread.User.Username,
		AvatarURL: thread.User.AvatarURL,
	}

	return &ThreadResponse{
		ID:           thread.ID,
		UserID:       thread.UserID,
		Title:        thread.Title,
		Content:      thread.Content,
		ImageURL:     thread.ImageURL,
		IsLiked:      isLiked,
		LikeCount:    thread.LikeCount,
		CommentCount: thread.CommentCount,
		User:         userBrief,
		CreatedAt:    thread.CreatedAt.Format(utils.TimeFormat),
		UpdatedAt:    thread.UpdatedAt.Format(utils.TimeFormat),
	}
}

func (s *ThreadService) checkIsLiked(userID, threadID uuid.UUID) bool {
	_, err := s.likeRepo.FindByUserAndThread(userID, threadID)
	return err == nil
}

func (s *ThreadService) Create(userID uuid.UUID, req CreateThreadRequest) (*ThreadResponse, error) {
	thread := &models.Thread{
		ID:       uuid.New(),
		UserID:   userID,
		Title:    req.Title,
		Content:  req.Content,
		ImageURL: req.ImageURL,
	}

	if err := s.threadRepo.Create(thread); err != nil {
		return nil, err
	}

	created, err := s.threadRepo.FindByID(thread.ID)
	if err != nil {
		return nil, err
	}

	return modelToThreadResponse(created, false), nil
}

func (s *ThreadService) GetAll(userID uuid.UUID, page, limit int) ([]ThreadResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	threads, total, err := s.threadRepo.FindAll(page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]ThreadResponse, len(threads))
	for i, t := range threads {
		isLiked := s.checkIsLiked(userID, t.ID)
		responses[i] = *modelToThreadResponse(&t, isLiked)
	}

	return responses, total, nil
}

func (s *ThreadService) GetByID(userID, id uuid.UUID) (*ThreadResponse, error) {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}

	isLiked := s.checkIsLiked(userID, thread.ID)
	return modelToThreadResponse(thread, isLiked), nil
}

func (s *ThreadService) Update(userID uuid.UUID, id uuid.UUID, req UpdateThreadRequest) (*ThreadResponse, error) {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrThreadNotFound
		}
		return nil, err
	}

	if thread.UserID != userID {
		return nil, ErrForbidden
	}

	thread.Title = req.Title
	thread.Content = req.Content
	thread.ImageURL = req.ImageURL

	if err := s.threadRepo.Update(thread); err != nil {
		return nil, err
	}

	updated, err := s.threadRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	isLiked := s.checkIsLiked(userID, updated.ID)
	return modelToThreadResponse(updated, isLiked), nil
}

func (s *ThreadService) Search(userID uuid.UUID, keyword string, page, limit int) ([]ThreadResponse, int64, error) {
	if keyword == "" {
		return []ThreadResponse{}, 0, nil
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	threads, total, err := s.threadRepo.Search(keyword, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]ThreadResponse, len(threads))
	for i, t := range threads {
		isLiked := s.checkIsLiked(userID, t.ID)
		responses[i] = *modelToThreadResponse(&t, isLiked)
	}

	return responses, total, nil
}

func (s *ThreadService) Delete(userID uuid.UUID, id uuid.UUID) error {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrThreadNotFound
		}
		return err
	}

	if thread.UserID != userID {
		return ErrForbidden
	}

	if err := s.threadRepo.Delete(id); err != nil {
		return err
	}

	// Cloudinary cleanup (best-effort, after DB success)
	if thread.ImageURL != "" {
		s.imageService.DeleteImage(thread.ImageURL)
	}

	return nil
}
