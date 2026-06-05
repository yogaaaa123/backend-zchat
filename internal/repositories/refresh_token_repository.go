package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *RefreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// FindByTokenHashAndDelete atomically finds and deletes a refresh token by its hash.
// Uses a transaction to prevent race conditions.
func (r *RefreshTokenRepository) FindByTokenHashAndDelete(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&token).Error; err != nil {
			return err
		}
		return tx.Delete(&models.RefreshToken{}, "id = ?", token.ID).Error
	})
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Delete(&models.RefreshToken{}, "id = ?", id).Error
}

func (r *RefreshTokenRepository) FindExpiredByHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Unscoped().
		Where("token_hash = ?", tokenHash).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *RefreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Delete(&models.RefreshToken{}, "user_id = ?", userID).Error
}

func (r *RefreshTokenRepository) DeleteExpired() error {
	return r.db.Delete(&models.RefreshToken{}, "expires_at < ?", time.Now()).Error
}
