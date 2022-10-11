package jwt

import "github.com/golang-jwt/jwt/v4"

type Claims struct {
	Type string
	Payload
	jwt.RegisteredClaims
}
