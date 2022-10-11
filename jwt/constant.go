package jwt

import (
	"time"
)

const (
	TypeKey         = "type"
	AccessTokenTTL  = time.Minute * 60 //TODO: lower value
	RefreshTokenTTL = time.Hour * (30 * 24)
)
