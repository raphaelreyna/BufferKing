package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Artist struct {
	Name   string
	Albums map[string]*Album
}

type Album struct {
	Name   string
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
}

func (l *Library) Stored(t *TrackSignal) bool {
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

func (l *Library) MarkStored(t *TrackSignal) {
	var ok bool

	// Does artist exist?
	var artist *Artist
	if artist, ok = l.Artists[t.Artist]; !ok {
		album := NewAlbum(t.Album, t.Title)

		l.Artists[t.Artist] = &Artist{
			Name:   t.Artist,
			Albums: map[string]*Album{t.Album: album},
		}

		return
	}

	// Does album exist?
	var album *Album
	if album, ok = artist.Albums[t.Album]; !ok {
		artist.Albums[t.Album] = NewAlbum(t.Album, t.Title)
		return
	}

	// Does track exist?
	if _, ok = album.Tracks[t.Title]; !ok {
		album.Tracks[t.Title] = struct{}{}
		return
	}
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
			return fmt.Errorf("invalid path: %s %+v", relPath, dirs)
		}
		// 0 -> TrackNumber; 1 -> Title; 2-> Format
		titleParts := re.FindStringSubmatch(dirs[2])

		tn, err := strconv.Atoi(titleParts[1])
		if err != nil {
			return err
		}

		track := &TrackSignal{
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
