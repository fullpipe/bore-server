package torrent

import (
	"fmt"

	"github.com/anacrolix/torrent"
	"github.com/fullpipe/bore-server/entity"
	"gorm.io/gorm"
)

func NewDownloader(dataDir string, db *gorm.DB) *Downloader {
	return &Downloader{
		dataDir: dataDir,
		db:      db,
	}
}

type Downloader struct {
	dataDir string
	db      *gorm.DB
}

func (dr *Downloader) Download(d *entity.Download) error {
	if d.State == entity.DownloadStateDone {
		return nil
	}

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = dr.dataDir
	cfg.NoUpload = true // TODO: make optional

	c, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}
	defer c.Close()
	fmt.Println(d)

	t, err := c.AddMagnet(d.Magnet)
	if err != nil {
		return err
	}
	<-t.GotInfo()
	// fmt.Println(t.Info().BestName())
	// fmt.Println(t.Info().Name)
	// fmt.Println(t.Info().NameUtf8)
	// fmt.Println(t.Info().Files[0])
	// fmt.Println(t.Name())
	// fmt.Println(t.Metainfo().Announce)
	// fmt.Println(t.Metainfo().Comment)

	// for _, f := range t.Files() {
	// 	f.SetPriority(torrent.PiecePriorityNone)
	// }

	// t.Files()[0].Download()

	// fmt.Println(t.Files()[0].DisplayPath())

	d.State = entity.DownloadStateInProgress
	d.Name = t.Info().BestName()
	dr.db.Save(&d)

	// TODO: handle download errors
	t.DownloadAll()

	done := c.WaitAll()
	if done {
		d.State = entity.DownloadStateDone
		dr.db.Save(&d)
	}

	return nil
}
