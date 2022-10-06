package entity

import "gorm.io/gorm"

type Book struct {
	gorm.Model

	DownloadID uint
	Download   *Download

	Parts []Part

	Title   string
	Author  string
	Reader  string
	Picture string
}
