package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/generated"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/fullpipe/bore-server/jwt"
	"github.com/fullpipe/bore-server/mail"
	"github.com/fullpipe/passhash"
	"gorm.io/gorm"
)

// Parts is the resolver for the parts field.
func (r *bookResolver) Parts(ctx context.Context, obj *entity.Book) ([]*entity.Part, error) {
	var parts []*entity.Part
	r.db.
		Where("book_id = ?", obj.ID).
		Order("possition ASC").
		Find(&parts)

	return parts, nil
}

// Download is the resolver for the download field.
func (r *bookResolver) Download(ctx context.Context, obj *entity.Book) (*entity.Download, error) {
	var d entity.Download
	r.db.
		First(&d, obj.DownloadID)

	return &d, nil
}

// CreateBook is the resolver for the createBook field.
func (r *mutationResolver) CreateBook(ctx context.Context, input model.NewBookInput) (*entity.Book, error) {
	// create download
	d := r.downloadRepo.FindByMagnet(input.Magnet)
	if d == nil {
		d = entity.NewDownload(input.Magnet)
	}
	r.db.Save(d)

	// create book
	book := r.bookRepo.FindByDownload(d.ID)
	if book == nil {
		book = &entity.Book{
			DownloadID: d.ID,
			State:      entity.BookStateDownload,
		}

		r.db.Save(book)
	}

	go r.downloadAndConvert(d, book)

	return book, nil
}

// LoginRequest is the resolver for the loginRequest field.
func (r *mutationResolver) LoginRequest(ctx context.Context, input model.LoginRequestInput) (uint, error) {
	// otp := utils.RandOTP()
	otp := "111112"
	hash, err := passhash.NewHash().HashPassword(otp)
	if err != nil {
		return 0, err
	}

	otpRequest := entity.LoginRequest{
		Email:     input.Email,
		Code:      hash,
		ExpiresAt: time.Now().Add(time.Minute * 100), //TODO: correct time
	}

	r.db.Create(&otpRequest)

	err = r.mailer.SendToEmail(
		"login.post_login_request",
		input.Email,
		mail.WithParam("otp", otp),
	)
	if err != nil {
		return 0, err
	}

	return otpRequest.ID, nil
}

// Login is the resolver for the login field.
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.Jwt, error) {
	var request entity.LoginRequest
	result := r.db.Where("expires_at > datetime()").Find(&request, input.RequestID)
	fmt.Println(result.Error)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("no request id")
	}

	if !passhash.NewHash().CheckPasswordHash(input.Code, request.Code) {
		return nil, errors.New("invalid code")
	}

	var user entity.User
	result = r.db.First(&user, &entity.User{Email: request.Email})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		user = entity.User{Email: request.Email}
		r.db.Save(&user)
	}

	r.db.Unscoped().Delete(&request)

	jwt, err := r.jwtBuilder.Build(jwt.Payload{
		UserID: user.ID,
		Roles:  user.Roles,
	})
	if err != nil {
		return nil, err
	}

	return &model.Jwt{
		Access:  jwt.AccessToken,
		Refresh: jwt.RefreshToken,
	}, nil
}

// Books is the resolver for the books field.
func (r *queryResolver) Books(ctx context.Context, filter *model.BooksFilter) ([]*entity.Book, error) {
	return r.bookRepo.All(), nil
}

// Book is the resolver for the book field.
func (r *queryResolver) Book(ctx context.Context, id uint) (*entity.Book, error) {
	book := r.bookRepo.FindByID(id)

	return book, nil
}

// Book returns generated.BookResolver implementation.
func (r *Resolver) Book() generated.BookResolver { return &bookResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type bookResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
