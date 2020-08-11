package library

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Artist struct {
	Name   string
	Albums map[string]*Album
}

type Album struct {
	Name string
	// Tracks holds track titles as keys for fast lookups
	// Track titles are formated as 'TrackNo - TrackTitle'
	Tracks map[string]struct{}
}

func NewAlbum(albumName, trackName string) *Album {
	return &Album{
		Name: albumName,
		Tracks: map[string]struct{}{
			trackName: struct{}{},
		},
	}
}

type Library struct {
	Root    string
	Artists map[string]*Artist
	sync.Mutex
}

func (l *Library) Stored(t *Track) bool {
	artists, ok := l.Artists[t.Artist]
	if !ok {
		return false
	}
	albums, ok := artists.Albums[t.Album]
	if !ok {
		return false
	}
	title := fmt.Sprintf("%d - %s", t.TrackNumber, t.Title)
	_, ok = albums.Tracks[title]
	return ok
}

func (l *Library) MarkStored(t *Track) {
	var ok bool
	title := fmt.Sprintf("%d - %s", t.TrackNumber, t.Title)

	// Does artist exist?
	var artist *Artist
	if artist, ok = l.Artists[t.Artist]; !ok {
		album := NewAlbum(t.Album, title)

		l.Artists[t.Artist] = &Artist{
			Name:   t.Artist,
			Albums: map[string]*Album{t.Album: album},
		}

		return
	}

	// Does album exist?
	var album *Album
	if album, ok = artist.Albums[t.Album]; !ok {
		artist.Albums[t.Album] = NewAlbum(t.Album, title)
		return
	}

	// Does track exist?
	if _, ok = album.Tracks[title]; !ok {
		album.Tracks[title] = struct{}{}
		return
	}
}

func (l *Library) UnmarkStored(t *Track) {
	var ok bool
	title := fmt.Sprintf("%d - %s", t.TrackNumber, t.Title)

	// Does artist exist?
	var artist *Artist
	if artist, ok = l.Artists[t.Artist]; !ok {
		return
	}

	// Does album exist?
	var album *Album
	if album, ok = artist.Albums[t.Album]; !ok {
		return
	}

	delete(album.Tracks, title)
}

// Unhide the file now that its finished
func (l *Library) FileMarkStored(t *Track, filename string) error {
	newPath := filepath.Join(l.Root, t.RelPath())

	dir := filepath.Dir(newPath)
	oldPath := filepath.Join(dir, filename)

	return os.Rename(oldPath, newPath)
}

func LoadLibrary(root string) (*Library, error) {
	l := &Library{Root: root, Artists: map[string]*Artist{}}
	re := regexp.MustCompile(`^([0-9]+) \- (.+)\.([0-9A-Za-z]+)$`)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}

		dirs := strings.Split(relPath, "/")
		if len(dirs) != 3 {
			// return fmt.Errorf("invalid path: %s %+v", relPath, dirs)
			return nil
		}
		// 0 -> TrackNumber; 1 -> Title; 2-> Format
		titleParts := re.FindStringSubmatch(dirs[2])

		if len(titleParts) < 3 {
			return nil
		}
		tn, err := strconv.Atoi(titleParts[1])
		if err != nil {
			return err
		}

		track := &Track{
			Artist:      dirs[0],
			Album:       dirs[1],
			Title:       titleParts[1] + " - " + titleParts[2],
			TrackNumber: int32(tn),
		}

		l.MarkStored(track)

		return nil
	})

	return l, err
}

func (l *Library) String() string {
	var s string

	newline := ""

	for _, artist := range l.Artists {
		s += newline
		s += fmt.Sprintf("Artist: %s\n", artist.Name)
		for _, album := range artist.Albums {
			s += fmt.Sprintf("\tAlbum: %s\n", album.Name)
			for track, _ := range album.Tracks {
				s += "\t\t" + track + "\n"
			}
		}
		newline = "\n"
	}

	return s
}
