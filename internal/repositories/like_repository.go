package repositories

import (
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"gorm.io/gorm"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) FindByUserAndThread(userID, threadID uuid.UUID) (*models.Like, error) {
	var like models.Like
	err := r.db.
		Where("user_id = ? AND thread_id = ?", userID, threadID).
		First(&like).Error
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *LikeRepository) Create(like *models.Like) error {
	return r.db.Create(like).Error
}

func (r *LikeRepository) Delete(id uuid.UUID) error {
	// Permanent delete (not soft) to avoid unique constraint conflict on re-like
	return r.db.Unscoped().Delete(&models.Like{}, "id = ?", id).Error
}

func (r *LikeRepository) CountByThread(threadID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("thread_id = ?", threadID).
		Count(&count).Error
	return count, err
}
