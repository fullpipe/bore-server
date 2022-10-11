package jwt

import (
	"crypto"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

type Parser interface {
	Parse(token string) (Payload, error)
}

type EdDSAParser struct {
	ed25519Key crypto.PrivateKey
	jwtType    string
}

func NewEdDSAParser(publicKey, jwtType string) (*EdDSAParser, error) {
	ed25519Key, err := jwt.ParseEdPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Ed25519 public key")
	}

	return &EdDSAParser{ed25519Key: ed25519Key, jwtType: jwtType}, nil
}

// Parse and validate token
func (p *EdDSAParser) Parse(tokenString string) (*Payload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		token.Method = jwt.SigningMethodEdDSA
		return p.ed25519Key, nil
	}, jwt.WithValidMethods([]string{"none"}))

	if err != nil {
		return nil, errors.Wrap(err, "unable to parse token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &claims.Payload, nil
	}

	return nil, errors.Wrap(err, "unable to parse claims")
}
