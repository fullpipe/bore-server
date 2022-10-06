package book

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fullpipe/bore-server/entity"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type Converter struct {
	bookDir string
}

func NewConverter(bookDir string) *Converter {
	return &Converter{bookDir: bookDir}
}

func (c *Converter) Convert(part entity.Part) error {
	c.bookDir = "./public/"

	dir := filepath.Join(c.bookDir, fmt.Sprintf("%d", part.BookID))
	out := filepath.Join(dir, fmt.Sprintf("%d.mp3", part.ID))

	os.MkdirAll(dir, 0777)

	err := ffmpeg.
		Input(part.Source).
		Output(out, ffmpeg.KwArgs{"q:a": 6, "vn": ""}).
		OverWriteOutput().ErrorToStdOut().Run()

	return err
}
