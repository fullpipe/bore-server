package entity

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email string
	Roles pq.StringArray `gorm:"type:text[]"`
}
