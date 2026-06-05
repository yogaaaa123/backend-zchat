package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thread struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	ImageURL  string         `gorm:"size:500" json:"image_url,omitempty"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Comments []Comment `gorm:"foreignKey:ThreadID" json:"comments,omitempty"`
	Likes    []Like    `gorm:"foreignKey:ThreadID" json:"likes,omitempty"`

	// Computed field (read-only, populated by subquery)
	LikeCount    int `gorm:"->" json:"like_count"`
	CommentCount int `gorm:"->" json:"comment_count"`
}

func (t *Thread) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *Thread) BeforeDelete(tx *gorm.DB) error {
	if t.ID != uuid.Nil {
		tx.Where("thread_id = ?", t.ID).Delete(&Comment{})
		tx.Where("thread_id = ?", t.ID).Delete(&Like{})
		tx.Where("thread_id = ?", t.ID).Delete(&Message{})
	}
	return nil
}
