package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/jwt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var userContextKey contextKey

type contextKey int

func JwtMiddleware(db *gorm.DB, jwtParser jwt.Parser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("authentication"), "Bearer ")
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			payload, err := jwtParser.Parse(token)
			if err != nil {
				http.Error(w, "Invalid access token", http.StatusUnauthorized)
			}

			log.Info("Payload: ", payload)

			ctx := context.WithValue(r.Context(), userContextKey, &userContext{
				db:      db,
				payload: payload,
			})

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

type userContext struct {
	db      *gorm.DB
	payload *jwt.Payload
	user    *entity.User
}

func User(ctx context.Context) *entity.User {
	uc, ok := ctx.Value(userContextKey).(*userContext)
	if !ok {
		return nil
	}

	if uc.user != nil {
		return uc.user
	}

	if uc.payload.UserID == 0 {
		return nil
	}

	var user entity.User
	result := uc.db.First(&user, uc.payload.UserID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	uc.user = &user

	return uc.user
}

func UserID(ctx context.Context) uint {
	raw, ok := ctx.Value(userContextKey).(*userContext)
	if !ok {
		return 0
	}

	return raw.payload.UserID
}

func Roles(ctx context.Context) []string {
	raw, ok := ctx.Value(userContextKey).(*userContext)
	if !ok {
		return []string{}
	}

	log.Info("Roles: ", raw.payload.Roles)

	return raw.payload.Roles
}
