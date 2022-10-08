package main

import (
	"log"
	"os"

	"github.com/fullpipe/bore-server/cmd/server"
	"github.com/fullpipe/bore-server/entity"
	"github.com/glebarez/sqlite"
	"github.com/urfave/cli"
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

	app := &cli.App{
		Name:  "boom",
		Usage: "make an explosive entrance",
		Commands: []cli.Command{
			server.NewCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
