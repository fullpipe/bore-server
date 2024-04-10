package entity

import (
	"fmt"

	"gorm.io/gorm"
)

type Part struct {
	gorm.Model

	BookID uint
	Book   *Book

	Title string

	// TODO: rename to position
	Possition uint

	Source string
	// Path   string

	Duration float64
}

func (p *Part) Path() string {
	return fmt.Sprintf("%d/%d.mp3", p.BookID, p.ID)
}
