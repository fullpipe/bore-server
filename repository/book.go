package repository

import (
	"errors"

	"github.com/fullpipe/bore-server/entity"
	"gorm.io/gorm"
)

type BookRepo struct {
	db *gorm.DB
}

func NewBookRepo(db *gorm.DB) *BookRepo {
	return &BookRepo{db}
}

func (r *BookRepo) All() []*entity.Book {
	var books []*entity.Book

	r.db.Find(&books)

	return books
}

func (r *BookRepo) FindWithProgress(userID uint) []*entity.Book {
	var books []*entity.Book

	r.db.Limit(10).
		Joins("INNER JOIN progress ON progress.book_id = book.id AND progress.user_id = ?", userID).
		Order("progress.updated_at DESC").Find(&books)

	return books
}

func (r *BookRepo) FindByID(bookID uint) *entity.Book {
	var b entity.Book

	result := r.db.Model(&entity.Book{}).First(&b, bookID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return &b
}

func (r *BookRepo) FindByDownload(downloadID uint) *entity.Book {
	var b entity.Book

	result := r.db.Model(&entity.Book{}).Where("download_id = ?", downloadID).First(&b)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return &b
}
