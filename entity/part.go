package entity

import "gorm.io/gorm"

type Part struct {
	gorm.Model

	BookID uint
	Book   *Book

	Title     string
	Possition uint

	Source string
	Path   string

	Duration uint
}
