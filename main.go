package main

import (
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
	"github.com/fullpipe/bore-server/repository"
	"github.com/fullpipe/bore-server/torrent"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const MagnetLink = "magnet:?xt=urn:btih:324C8EA62583CB95FA59A6522C1E132813CE5AB8&tr=http%3A%2F%2Fbt2.t-ru.org%2Fann%3Fmagnet&dn=%D0%9A%D1%80%D0%B0%D0%BF%D0%B8%D0%B2%D0%B8%D0%BD%20%D0%92%D0%BB%D0%B0%D0%B4%D0%B8%D1%81%D0%BB%D0%B0%D0%B2%20-%20%D0%94%D0%B5%D1%82%D1%81%D0%BA%D0%B0%D1%8F%20%D0%B0%D1%83%D0%B4%D0%B8%D0%BE%D0%BA%D0%BD%D0%B8%D0%B3%D0%B0%2C%20%D0%94%D0%B5%D1%82%D0%B8%20%D1%81%D0%B8%D0%BD%D0%B5%D0%B3%D0%BE%20%D1%84%D0%BB%D0%B0%D0%BC%D0%B8%D0%BD%D0%B3%D0%BE%20%5B%D0%A7%D0%BE%D0%B2%D0%B6%D0%B8%D0%BA%20%D0%90%D0%BB%D0%BB%D0%B0%2C%202019%2C%2064%20kbps%2C%20MP3%5D"

func main() {
	db, err := gorm.Open(sqlite.Open("lite.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&entity.Download{})
	db.AutoMigrate(&entity.Book{})
	db.AutoMigrate(&entity.Part{})

	// create download
	drp := repository.NewDownloadRepo(db)
	d := drp.FindByMagnet(MagnetLink)
	fmt.Println(d)
	if d == nil {
		d = entity.NewDownload(MagnetLink)
	}
	db.Save(d)

	// start download
	downloader := torrent.NewDownloader("./downloads", db)
	err = downloader.Download(d)
	if err != nil {
		log.Fatalln(err)
	}

	// create book
	book := &entity.Book{
		DownloadID: d.ID,
		Title:      d.Name,
	}
	db.Save(book)

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

		db.Save(part)

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

	bkr := repository.NewBookRepo(db)
	book = bkr.FindByID(book.ID)
	converter := bookSrv.NewConverter("./public")
	for _, part := range book.Parts {
		err := converter.Convert(part)
		if err != nil {
			log.Fatal(err)
		}
	}

	// book ready
	// remove download
}

func convert(part *entity.Part) error {
	fmt.Println(*part)
	return nil
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
		return strings.Compare(paths[i], paths[j]) > 0
	})

	return paths, nil
}

// func filesInOrder(files []fs.DirEntry) []fs.DirEntry {
// 	inOrder := []fs.DirEntry{}
// 	sort.Slice(files, func(i, j int) bool {
// 		return strings.Compare(files[i].Name(), files[j].Name()) > 0
// 	})

// 	for _, f := range files {
// 		if f.IsDir() {
// 			inOrder = append(inOrder, filesInOrder([]fs.DirEntry{f})...)
// 		} else {
// 			inOrder = append(inOrder, f)
// 		}
// 	}

// 	return inOrder
// }
