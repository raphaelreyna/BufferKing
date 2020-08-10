package library

import (
	"fmt"
	"path/filepath"
	"time"
)

type Track struct {
	Title       string
	Artist      string
	Album       string
	TrackNumber int32
	Length      time.Duration
	Format      string
}

func (t *Track) RelPath() string {
	var ext string
	if t.Format != "" {
		ext = "." + t.Format
	}
	return filepath.Join(t.Artist, t.Album, fmt.Sprintf("%d - %s%s", t.TrackNumber, t.Title, ext))
}

func (t *Track) String() string {
	return fmt.Sprintf("%s\t%s\t%s", t.Artist, t.Album, t.Title)
}
