package repository

import (
	"errors"

	"github.com/fullpipe/bore-server/entity"
	"gorm.io/gorm"
)

type ProgressRepo struct {
	db *gorm.DB
}

func NewProgressRepo(db *gorm.DB) *ProgressRepo {
	return &ProgressRepo{db}
}

func (r *ProgressRepo) FindByBook(bookID uint, userID uint) *entity.Progress {
	var p entity.Progress

	result := r.db.Where("user_id = ? AND book_id = ?", userID, bookID).First(&p)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return &p
}
