package repository

import (
	"errors"
	"fmt"

	"github.com/fullpipe/bore-server/entity"
	"gorm.io/gorm"
)

type DownloadRepo struct {
	db *gorm.DB
}

func NewDownloadRepo(db *gorm.DB) *DownloadRepo {
	return &DownloadRepo{db}
}

func (r *DownloadRepo) FindByMagnet(magnet string) *entity.Download {
	var d entity.Download

	result := r.db.Where("magnet = ?", magnet).First(&d)
	fmt.Println(result.Error)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("NOT FOUND")
		return nil
	}

	return &d
}
