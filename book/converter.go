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
	return &Converter{
		bookDir: bookDir,
	}
}

func (c *Converter) Convert(part entity.Part) error {
	dir := filepath.Join(c.bookDir, fmt.Sprintf("%d", part.BookID))
	out := filepath.Join(dir, fmt.Sprintf("%d.mp3", part.ID))

	os.MkdirAll(dir, 0777)

	err := ffmpeg.
		Input(part.Source).
		// Output(out, ffmpeg.KwArgs{"threads:v": 1, "q:a": 4, "vn": ""}). // https://trac.ffmpeg.org/wiki/Encode/MP3#VBREncoding
		Output(out, ffmpeg.KwArgs{"q:a": 4}). // https://trac.ffmpeg.org/wiki/Encode/MP3#VBREncoding
		OverWriteOutput().ErrorToStdOut().Run()

	return err
}

func (c *Converter) Delete(book *entity.Book) error {
	dir := filepath.Join(c.bookDir, fmt.Sprintf("%d", book.ID))

	return os.RemoveAll(dir)
}
