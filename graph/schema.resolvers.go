package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fullpipe/bore-server/auth"
	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/generated"
	"github.com/fullpipe/bore-server/graph/model"
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

// Progress is the resolver for the progress field.
func (r *bookResolver) Progress(ctx context.Context, obj *entity.Book) (*entity.Progress, error) {
	user := auth.User(ctx)
	if user == nil {
		return nil, nil
	}

	return r.progressRepo.FindByBook(obj.ID, user.ID), nil
}

// CreateBook is the resolver for the createBook field.
func (r *mutationResolver) CreateBook(ctx context.Context, input model.NewBookInput) (*entity.Book, error) {
	// create download
	d := r.downloadRepo.FindByMagnet(input.Magnet)
	if d == nil {
		d = entity.NewDownload(input.Magnet)
		r.db.Save(d)
	}

	// create book
	book := r.bookRepo.FindByDownload(d.ID)
	if book == nil {
		book = &entity.Book{
			DownloadID: d.ID,
			State:      entity.BookStateDownload,
		}

		// reinit download if its a new book
		d.State = entity.DownloadStateNew
		r.db.Save(d)

		r.db.Save(book)
	}

	go r.downloadAndConvert(d, book)

	return book, nil
}

// Delete is the resolver for the delete field.
func (r *mutationResolver) Delete(ctx context.Context, bookID uint) (bool, error) {
	book := r.bookRepo.FindByID(bookID)
	if book == nil {
		return false, errors.New("book not found")
	}

	err := r.converter.Delete(book)
	if err != nil {
		return false, err
	}
	r.db.Delete(book)

	download := r.downloadRepo.FindByID(book.DownloadID)
	if download == nil {
		return false, errors.New("book not found")
	}

	err = r.downloader.Delete(download)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Restart is the resolver for the restart field.
func (r *mutationResolver) Restart(ctx context.Context, bookID uint) (bool, error) {
	panic(fmt.Errorf("not implemented: Restart - restart"))
}

// RefreshToken is the resolver for the refreshToken field.
func (r *mutationResolver) RefreshToken(ctx context.Context, refreshToken string) (*model.Jwt, error) {
	payload, err := r.refreshParser.Parse(refreshToken)
	if err != nil {
		return nil, err
	}

	var user entity.User
	result := r.db.First(&user, payload.UserID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not exists")
	}

	return jwtResponce(r.jwtBuilder, &user)
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
	email := strings.ToLower(request.Email)
	result = r.db.First(&user, &entity.User{Email: email})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		user = entity.User{Email: email}
		r.db.Save(&user)
	}

	r.db.Unscoped().Delete(&request)

	return jwtResponce(r.jwtBuilder, &user)
}

// Progress is the resolver for the progress field.
func (r *mutationResolver) Progress(ctx context.Context, input model.ProgressInput) (*entity.Progress, error) {
	user := auth.User(ctx)
	if user == nil {
		return nil, errors.New("auth required")
	}

	p := r.progressRepo.FindByBook(input.BookID, user.ID)
	if p == nil {
		p = &entity.Progress{
			BookID: input.BookID,
			UserID: user.ID,
		}
	}

	p.Part = input.Part
	p.Speed = input.Speed
	p.Position = input.Position
	p.GlobalDuration = input.GlobalDuration
	p.GlobalPosition = input.GlobalPosition

	r.db.Save(p)

	return p, nil
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

// LastBooks is the resolver for the lastBooks field.
func (r *queryResolver) LastBooks(ctx context.Context) ([]*entity.Book, error) {
	user := auth.User(ctx)
	if user == nil {
		return nil, nil
	}

	return r.bookRepo.FindWithProgress(user.ID), nil
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
