package graph

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	bookSrv "github.com/fullpipe/bore-server/book"
	"github.com/fullpipe/bore-server/config"
	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/fullpipe/bore-server/jwt"
	"github.com/fullpipe/bore-server/mail"
	"github.com/fullpipe/bore-server/repository"
	"github.com/fullpipe/bore-server/torrent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vansante/go-ffprobe.v2"
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

		partTitle := filepath.Base(f.Name())
		m, err := tag.ReadFrom(f)
		if err != nil {
			log.Error(err)
		} else {
			if m.Album() != "" {
				book.Title = m.Album()
			}

			if m.Artist() != "" && book.Author == "" {
				book.Author = m.Artist()
			}

			if m.AlbumArtist() != "" && book.Reader == "" {
				book.Reader = m.AlbumArtist()
			}
			if m.Title() != "" {
				partTitle = strings.TrimSpace(m.Title())
			}
		}

		// TODO: add book piture
		part := &entity.Part{
			BookID:    book.ID,
			Title:     partTitle,
			Possition: uint(i),
			Source:    path,
		}

		part.Duration, err = getFileDuration(context.Background(), part.Source)
		if err != nil {
			logrus.Error(err)
		}

		r.db.Save(part)

		parts = append(parts, part)
	}

	if book.Title == "" {
		book.Title = d.Name
	}

	book.State = entity.BookStateConvert
	r.db.Save(book)

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

func getFileDuration(ctx context.Context, path string) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := ffprobe.ProbeURL(ctx, path)
	if err != nil {
		return 0, errors.Wrap(err, "getFileDuration")
	}

	return data.Format.DurationSeconds, nil
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
