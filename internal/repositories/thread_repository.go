package repositories

import (
	"github.com/google/uuid"
	"github.com/satria/obrolan-api/internal/models"
	"gorm.io/gorm"
)

type ThreadRepository struct {
	db *gorm.DB
}

func NewThreadRepository(db *gorm.DB) *ThreadRepository {
	return &ThreadRepository{db: db}
}

func (r *ThreadRepository) Create(thread *models.Thread) error {
	return r.db.Create(thread).Error
}

func (r *ThreadRepository) FindByID(id uuid.UUID) (*models.Thread, error) {
	var thread models.Thread
	err := 	r.db.
		Preload("User").
		Select("threads.*, (SELECT COUNT(*) FROM likes WHERE likes.thread_id = threads.id) AS like_count, (SELECT COUNT(*) FROM comments WHERE comments.thread_id = threads.id) AS comment_count").
		First(&thread, "threads.id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *ThreadRepository) FindAll(page, limit int) ([]models.Thread, int64, error) {
	var threads []models.Thread
	var total int64

	db := 	r.db.Model(&models.Thread{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := 	r.db.
		Preload("User").
		Select("threads.*, (SELECT COUNT(*) FROM likes WHERE likes.thread_id = threads.id) AS like_count, (SELECT COUNT(*) FROM comments WHERE comments.thread_id = threads.id) AS comment_count").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads).Error
	if err != nil {
		return nil, 0, err
	}

	return threads, total, nil
}

func (r *ThreadRepository) Update(thread *models.Thread) error {
	return r.db.Model(thread).Updates(map[string]interface{}{
		"title":     thread.Title,
		"content":   thread.Content,
		"image_url": thread.ImageURL,
	}).Error
}

func (r *ThreadRepository) Search(keyword string, page, limit int) ([]models.Thread, int64, error) {
	var threads []models.Thread
	var total int64

	query := "%" + keyword + "%"
	likeClause := "LOWER(title) LIKE LOWER(?) OR LOWER(content) LIKE LOWER(?)"

	db := 	r.db.Model(&models.Thread{}).Where(likeClause, query, query)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := 	r.db.
		Where(likeClause, query, query).
		Preload("User").
		Select("threads.*, (SELECT COUNT(*) FROM likes WHERE likes.thread_id = threads.id) AS like_count, (SELECT COUNT(*) FROM comments WHERE comments.thread_id = threads.id) AS comment_count").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threads).Error
	if err != nil {
		return nil, 0, err
	}

	return threads, total, nil
}

func (r *ThreadRepository) Delete(id uuid.UUID) error {
	// Use model with ID so BeforeDelete hook gets the populated model
	return r.db.Delete(&models.Thread{ID: id}).Error
}

func (r *ThreadRepository) FindThreadExists(threadID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Thread{}).Where("id = ?", threadID).Count(&count).Error
	return count > 0, err
}
