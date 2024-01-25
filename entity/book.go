package entity

import "gorm.io/gorm"

const (
	BookStateDownload BookState = "download"
	BookStateError    BookState = "error"
	BookStateConvert  BookState = "convert"
	BookStateReady    BookState = "ready"
)

type Book struct {
	gorm.Model

	DownloadID uint
	Download   *Download

	State BookState

	Title   string
	Author  string
	Reader  string
	Picture string
	Error   string
}

type BookState string
