//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"os"

	"github.com/dhowden/tag"
	bookSrv "github.com/fullpipe/bore-server/book"
	"github.com/fullpipe/bore-server/config"
	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/fullpipe/bore-server/jwt"
	"github.com/fullpipe/bore-server/mail"
	"github.com/fullpipe/bore-server/repository"
	"github.com/fullpipe/bore-server/torrent"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Resolver struct {
	db            *gorm.DB
	bookRepo      *repository.BookRepo
	downloadRepo  *repository.DownloadRepo
	progressRepo  *repository.ProgressRepo
	downloader    *torrent.Downloader
	converter     *bookSrv.Converter
	mailer        *mail.Mailer
	jwtBuilder    *jwt.Builder
	refreshParser jwt.Parser
}

func NewResolver(
	db *gorm.DB,
	cfg config.Config,

) *Resolver {
	mailer, _ := mail.NewMailer(cfg.Mailer)
	signer, _ := jwt.NewEdDSASigner(cfg.JWT.PrivateKey)
	jwtBuilder := jwt.NewBuilder(signer)
	refreshParser, _ := jwt.NewEdDSAParser(cfg.JWT.PublicKey, "refresh")

	return &Resolver{
		db:            db,
		bookRepo:      repository.NewBookRepo(db),
		downloadRepo:  repository.NewDownloadRepo(db),
		progressRepo:  repository.NewProgressRepo(db),
		downloader:    torrent.NewDownloader(cfg.TorrentsDir, db),
		converter:     bookSrv.NewConverter(cfg.BooksDir),
		mailer:        mailer,
		jwtBuilder:    jwtBuilder,
		refreshParser: refreshParser,
	}
}

func (r *mutationResolver) downloadAndConvert(d *entity.Download, book *entity.Book) {
	if book.State == entity.BookStateReady {
		log.Infof("Book #%d %s already downloaded", book.ID, book.Title)
		return
	}

	// start download
	err := r.downloader.Download(d)
	if err != nil {
		log.Error(err)
	}

	// get downloaded files in order
	paths, err := r.downloader.GetFilePathsInOrder(d)
	if err != nil {
		log.Error(err)
	}

	parts := []*entity.Part{}
	for i, path := range paths {
		f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Error(err)
		}

		m, err := tag.ReadFrom(f)
		if err != nil {
			log.Error(err)
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
			book.State = entity.BookStateError
			r.db.Save(book)
			log.Error(err)
			return
		}
	}

	book.State = entity.BookStateReady
	book.Error = ""

	r.db.Save(book)

	err = r.downloader.Delete(d)
	if err != nil {
		log.Error(err)
	}
}

func jwtResponce(jwtBuilder *jwt.Builder, user *entity.User) (*model.Jwt, error) {
	jwt, err := jwtBuilder.Build(jwt.Payload{
		UserID: user.ID,
		Roles:  user.Roles,
	})

	if err != nil {
		return nil, err
	}

	roles := []model.Role{}
	for _, r := range user.Roles {
		roles = append(roles, model.Role(r))
	}

	return &model.Jwt{
		Access:  jwt.AccessToken,
		Refresh: jwt.RefreshToken,
		Roles:   roles,
	}, nil
}
