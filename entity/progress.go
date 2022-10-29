package entity

import "gorm.io/gorm"

type Progress struct {
	gorm.Model

	BookID uint
	Book   *Book

	UserID uint
	User   *User

	Part           uint
	Speed          float64
	Position       float64
	GlobalDuration float64
	GlobalPosition float64
}
