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

func (r *BookRepo) FindByID(bookID uint) *entity.Book {
	var b entity.Book

	r.db.Model(&entity.Book{}).First(&b, bookID)

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
