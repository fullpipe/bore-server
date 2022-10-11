package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Builder struct {
	signer Signer
}

type Pair struct {
	RefreshToken string
	AccessToken  string
	IssuedAt     time.Time
}

func NewBuilder(signer Signer) *Builder {
	return &Builder{signer: signer}
}

func (b *Builder) Build(payload Payload) (Pair, error) {
	issuedAt := time.Now().UTC().Truncate(time.Second)

	claims := Claims{
		Type:    "access",
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(issuedAt),
			Issuer:   "test",
		},
	}

	claims.Type = "access"
	claims.ExpiresAt = jwt.NewNumericDate(issuedAt.Add(AccessTokenTTL))
	accessStr, err := b.signer.Sign(claims)
	if err != nil {
		return Pair{}, err
	}

	claims.Type = "refresh"
	claims.ExpiresAt = jwt.NewNumericDate(issuedAt.Add(RefreshTokenTTL))
	refreshStr, err := b.signer.Sign(claims)
	if err != nil {
		return Pair{}, err
	}

	return Pair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		IssuedAt:     issuedAt,
	}, nil
}
