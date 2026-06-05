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
	ErrCommentNotFound = errors.New("comment not found")
	ErrParentNotFound  = errors.New("parent comment not found")
	ErrInvalidParent   = errors.New("parent comment does not belong to this thread")
	ErrInvalidParentID = errors.New("invalid parent_id format")
)

type CommentService struct {
	commentRepo *repositories.CommentRepository
}

func NewCommentService(commentRepo *repositories.CommentRepository) *CommentService {
	return &CommentService{commentRepo: commentRepo}
}

type CreateCommentRequest struct {
	Content  string   `json:"content" binding:"required,min=1"`
	ParentID *string  `json:"parent_id"` // optional UUID string for nested replies
}

type CommentUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

type CommentResponse struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	ThreadID  uuid.UUID           `json:"thread_id"`
	ParentID  *uuid.UUID          `json:"parent_id"`
	Content   string              `json:"content"`
	User      *CommentUserResponse `json:"user,omitempty"`
	Replies   []CommentResponse   `json:"replies,omitempty"`
	CreatedAt string              `json:"created_at"`
}

func modelToCommentResponse(comment models.Comment) CommentResponse {
	resp := CommentResponse{
		ID:        comment.ID,
		UserID:    comment.UserID,
		ThreadID:  comment.ThreadID,
		ParentID:  comment.ParentID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(utils.TimeFormat),
	}

	if comment.User.ID != uuid.Nil {
		resp.User = &CommentUserResponse{
			ID:        comment.User.ID,
			Username:  comment.User.Username,
			AvatarURL: comment.User.AvatarURL,
		}
	}

	// Map nested replies
	if len(comment.Replies) > 0 {
		replies := make([]CommentResponse, len(comment.Replies))
		for i, reply := range comment.Replies {
			replyResp := modelToCommentResponse(reply)
			replyResp.Replies = nil // prevent deep nesting beyond 1 level
			replies[i] = replyResp
		}
		resp.Replies = replies
	}

	return resp
}

func (s *CommentService) Create(userID uuid.UUID, threadID uuid.UUID, req CreateCommentRequest) (*CommentResponse, error) {
	exists, err := s.commentRepo.FindThreadExists(threadID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrThreadNotFound
	}

	comment := &models.Comment{
		ID:       uuid.New(),
		UserID:   userID,
		ThreadID: threadID,
		Content:  req.Content,
	}

	if req.ParentID != nil && *req.ParentID != "" {
		parentUUID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, ErrInvalidParentID
		}

		parent, err := s.commentRepo.FindByID(parentUUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentNotFound
			}
			return nil, err
		}

		if parent.ThreadID != threadID {
			return nil, ErrInvalidParent
		}

		comment.ParentID = &parentUUID
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	savedComment, err := s.commentRepo.FindByIDWithUser(comment.ID)
	if err != nil {
		return nil, err
	}

	resp := modelToCommentResponse(*savedComment)
	resp.Replies = nil
	return &resp, nil
}

func (s *CommentService) GetByThread(threadID uuid.UUID, page, limit int) ([]CommentResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	exists, err := s.commentRepo.FindThreadExists(threadID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, ErrThreadNotFound
	}

	comments, total, err := s.commentRepo.FindByThreadID(threadID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]CommentResponse, len(comments))
	for i, c := range comments {
		responses[i] = modelToCommentResponse(c)
	}

	return responses, total, nil
}

func (s *CommentService) Delete(userID uuid.UUID, commentID uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID {
		return ErrForbidden
	}

	if err := s.commentRepo.Delete(commentID); err != nil {
		return err
	}

	if err := s.commentRepo.DeleteRepliesByParentID(commentID); err != nil {
		return err
	}

	return nil
}
