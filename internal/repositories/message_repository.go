package repositories

import (
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) GetByThread(threadID uuid.UUID, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.
		Preload("User").
		Where("thread_id = ?", threadID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
