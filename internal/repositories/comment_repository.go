package repositories

import (
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) FindByThreadID(threadID uuid.UUID, page, limit int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	// Count total top-level comments
	r.db.Model(&models.Comment{}).
		Where("thread_id = ? AND parent_id IS NULL", threadID).
		Count(&total)

	// Get paginated top-level comments with user + nested replies
	offset := (page - 1) * limit
	err := r.db.
		Where("thread_id = ? AND parent_id IS NULL", threadID).
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Replies.User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *CommentRepository) FindByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "id = ?", id).Error
}

func (r *CommentRepository) FindThreadExists(threadID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Thread{}).Where("id = ?", threadID).Count(&count).Error
	return count > 0, err
}

func (r *CommentRepository) FindByIDWithUser(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("User").First(&comment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) DeleteRepliesByParentID(parentID uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "parent_id = ?", parentID).Error
}
