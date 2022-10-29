package entity

import "gorm.io/gorm"

const (
	DownloadStateNew        DownloadState = "new"
	DownloadStateInProgress DownloadState = "in_progress"
	DownloadStateError      DownloadState = "error"
	DownloadStateDone       DownloadState = "done"
	DownloadStateDelete     DownloadState = "delete"
)

func NewDownload(magnet string) *Download {
	return &Download{
		Magnet: magnet,
		State:  DownloadStateNew,
	}
}

type Download struct {
	gorm.Model

	Name   string
	Magnet string `gorm:"index"`
	State  DownloadState

	Length     int64
	Downloaded int64
	Error      string
}

type DownloadState string
