//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dhowden/tag"
	bookSrv "github.com/fullpipe/bore-server/book"
	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/fullpipe/bore-server/repository"
	"github.com/fullpipe/bore-server/torrent"
	"gorm.io/gorm"
)

type Resolver struct {
	db       *gorm.DB
	bookRepo *repository.BookRepo
}

func NewResolver(
	db *gorm.DB,
	bookRepo *repository.BookRepo,

) *Resolver {
	return &Resolver{
		db:       db,
		bookRepo: bookRepo,
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

// CreateBook is the resolver for the createBook field.
func (r *mutationResolver) CreateBook(ctx context.Context, input model.NewBook) (*entity.Book, error) {
	// create download
	drp := repository.NewDownloadRepo(r.db)
	d := drp.FindByMagnet(input.Magnet)
	fmt.Println(d)
	if d == nil {
		d = entity.NewDownload(input.Magnet)
	}
	r.db.Save(d)

	// start download
	downloader := torrent.NewDownloader("./downloads", r.db)
	err := downloader.Download(d)
	if err != nil {
		log.Fatalln(err)
	}

	// create book
	book := &entity.Book{
		DownloadID: d.ID,
		Title:      d.Name,
	}
	r.db.Save(book)

	// get downloaded files in order
	paths, err := getFilePathsInOrder(d)
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

	// get meta info
	// create BookPart
	// 	source = source file path
	//  dest = destination file path
	//  meta
	//  duration
	// 	possition = i
	//
	// convert them to webp

	converter := bookSrv.NewConverter("./public")
	for _, part := range parts {
		err := converter.Convert(*part)
		if err != nil {
			log.Fatal(err)
		}
	}

	return book, nil
}

func getFilePathsInOrder(d *entity.Download) ([]string, error) {
	root := path.Join("./downloads", d.Name)
	paths := []string{}

	err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				paths = append(paths, path)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	sort.Slice(paths, func(i, j int) bool {
		return strings.Compare(paths[i], paths[j]) < 0
	})

	return paths, nil
}
