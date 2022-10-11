//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"context"
	"log"
	"os"

	"github.com/dhowden/tag"
	bookSrv "github.com/fullpipe/bore-server/book"
	"github.com/fullpipe/bore-server/config"
	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/fullpipe/bore-server/repository"
	"github.com/fullpipe/bore-server/torrent"
	"gorm.io/gorm"
)

type Resolver struct {
	db           *gorm.DB
	bookRepo     *repository.BookRepo
	downloadRepo *repository.DownloadRepo
	downloader   *torrent.Downloader
	converter    *bookSrv.Converter
}

func NewResolver(
	db *gorm.DB,
	cfg config.Config,

) *Resolver {
	return &Resolver{
		db:           db,
		bookRepo:     repository.NewBookRepo(db),
		downloadRepo: repository.NewDownloadRepo(db),
		downloader:   torrent.NewDownloader(cfg.TorrentsDir, db),
		converter:    bookSrv.NewConverter(cfg.BooksDir),
	}
}

// Book is the resolver for the book field.
func (r *queryResolver) Book(ctx context.Context, id uint) (*entity.Book, error) {
	book := r.bookRepo.FindByID(id)

	return book, nil
}

// Books is the resolver for the books field.
func (r *queryResolver) Books(ctx context.Context, filter *model.BooksFilter) ([]*entity.Book, error) {
	return r.bookRepo.All(), nil
}

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
func (r *bookResolver) Download(ctx context.Context, book *entity.Book) (*entity.Download, error) {
	var d entity.Download
	r.db.
		First(&d, book.DownloadID)

	return &d, nil
}

// CreateBook is the resolver for the createBook field.
func (r *mutationResolver) CreateBook(ctx context.Context, input model.NewBook) (*entity.Book, error) {
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
		}

		r.db.Save(book)
	}

	go r.downloadAndConvert(d, book)

	return book, nil
}

func (r *mutationResolver) downloadAndConvert(d *entity.Download, book *entity.Book) {
	// start download
	err := r.downloader.Download(d)
	if err != nil {
		log.Fatalln(err)
	}

	// get downloaded files in order
	paths, err := r.downloader.GetFilePathsInOrder(d)
	if err != nil {
		log.Fatalln(err)
	}

	parts := []*entity.Part{}
	for i, path := range paths {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatalln(err)
		}

		m, err := tag.ReadFrom(f)
		if err != nil {
			log.Fatal(err)
		}

		if m.Album() != "" {
			book.Title = m.Album()
		}

		if m.Artist() != "" && book.Author == "" {
			book.Author = m.Artist()
		}

		if m.AlbumArtist() != "" && book.Reader == "" {
			book.Reader = m.AlbumArtist()
		}

		// TODO: add book piture
		// book.Picture: "???",

		part := &entity.Part{
			BookID:    book.ID,
			Title:     m.Title(),
			Possition: uint(i),
			Source:    path,
		}

		r.db.Save(part)

		parts = append(parts, part)
	}

	r.db.Save(book)

	// get meta info
	// create BookPart
	// 	source = source file path
	//  dest = destination file path
	//  meta
	//  duration
	// 	possition = i
	//
	// convert them to webp

	for _, part := range parts {
		err := r.converter.Convert(*part)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = r.downloader.Delete(d)
	if err != nil {
		log.Fatal(err)
	}
}
