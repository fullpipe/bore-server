package entity

import "gorm.io/gorm"

type Progress struct {
	gorm.Model

	BookID uint
	Book   *Book

	UserID uint
	User   *User

	Part     uint
	Position float64
	Speed    float64
}
