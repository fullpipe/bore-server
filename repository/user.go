package repository

import (
	"errors"

	"github.com/fullpipe/bore-server/entity"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db}
}

func (r *UserRepo) FindByID(userID uint) *entity.User {
	var u entity.User

	result := r.db.Model(&entity.Book{}).First(&u, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	return &u
}
