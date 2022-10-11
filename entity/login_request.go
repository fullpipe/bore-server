package entity

import (
	"time"

	"gorm.io/gorm"
)

type LoginRequest struct {
	gorm.Model

	Email     string
	Code      string
	ExpiresAt time.Time
}
