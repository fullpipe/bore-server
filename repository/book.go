package repository

import (
	"github.com/fullpipe/bore-server/entity"
	"gorm.io/gorm"
)

type BookRepo struct {
	db *gorm.DB
}

func NewBookRepo(db *gorm.DB) *BookRepo {
	return &BookRepo{db}
}

func (r *BookRepo) FindByID(bookID uint) *entity.Book {
	var b entity.Book

	r.db.Model(&entity.Book{}).Preload("Parts").First(&b, bookID)

	return &b
}
