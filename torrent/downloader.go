package torrent

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/fullpipe/bore-server/entity"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func NewDownloader(dataDir string, db *gorm.DB) *Downloader {
	os.MkdirAll(dataDir, 0777)

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
	cfg.ListenPort = 0
	cfg.NoUpload = true // TODO: make optional

	c, err := torrent.NewClient(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to start torrent client")
	}
	defer c.Close()

	t, err := c.AddMagnet(d.Magnet)
	if err != nil {
		return errors.Wrap(err, "unable to get info")
	}
	<-t.GotInfo()

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

func (dr *Downloader) GetFilePathsInOrder(d *entity.Download) ([]string, error) {
	root := path.Join(dr.dataDir, d.Name)
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

func (dr *Downloader) Delete(d *entity.Download) error {
	root := path.Join(dr.dataDir, d.Name)
	err := os.RemoveAll(root)
	if err != nil {
		return errors.Wrap(err, "can not delete download")
	}

	d.State = entity.DownloadStateDelete

	return dr.db.Save(&d).Error
}
