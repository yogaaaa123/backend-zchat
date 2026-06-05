package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ThreadID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"thread_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Thread Thread `gorm:"foreignKey:ThreadID" json:"thread,omitempty"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
