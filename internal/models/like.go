package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Like struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_thread;index" json:"user_id"`
	ThreadID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_thread;index" json:"thread_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Thread Thread `gorm:"foreignKey:ThreadID" json:"thread,omitempty"`
}

func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
