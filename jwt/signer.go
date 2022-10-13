package jwt

import (
	"crypto"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

type Signer interface {
	Sign(claims Claims) (string, error)
}

type EdDSASigner struct {
	ed25519Key crypto.PrivateKey
}

func NewEdDSASigner(privateKey string) (Signer, error) {
	ed25519Key, err := jwt.ParseEdPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Ed25519 private key")
	}

	return &EdDSASigner{ed25519Key: ed25519Key}, nil
}

func (s *EdDSASigner) Sign(claims Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["alg"] = "none"
	return token.SignedString(s.ed25519Key)
}
