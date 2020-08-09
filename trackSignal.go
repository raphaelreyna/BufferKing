package main

import (
	"fmt"
	dbus "github.com/godbus/dbus/v5"
	"path/filepath"
	"time"
)

type Status int

const (
	Play  Status = 1
	Pause Status = -1

	// Used for comparing Track signals; diff = before - after
	None     Status = Play - Play  // 0
	Paused   Status = Play - Pause // 2
	Resumed  Status = Pause - Play // -2
	NewTrack Status = 3
)

type TrackSignal struct {
	Title       string
	Artist      string
	Album       string
	TrackNumber int32
	Length      time.Duration
	Stat        Status
}

func ParseSignal(sign *dbus.Signal) (*TrackSignal, error) {
	resp := map[string]dbus.Variant{}
	err := dbus.Store(sign.Body[1:2], &resp)
	if err != nil {
		return nil, err
	}

	var stat Status
	switch sstat := resp["PlaybackStatus"].Value().(string); sstat {
	case "Playing":
		stat = Play
	case "Paused":
		stat = Pause
	}

	md := resp["Metadata"].Value().(map[string]dbus.Variant)

	return &TrackSignal{
		Title:       md["xesam:title"].Value().(string),
		Artist:      md["xesam:artist"].Value().([]string)[0],
		Album:       md["xesam:album"].Value().(string),
		TrackNumber: md["xesam:trackNumber"].Value().(int32),
		Length:      time.Duration(md["mpris:length"].Value().(uint64)) * time.Microsecond,
		Stat:        stat,
	}, nil
}

// Compare compares the old tracksignal to the new tracksignal
func (t *TrackSignal) Compare(tt *TrackSignal) Status {
	if t == nil && tt != nil {
		return NewTrack
	}
	sameTrack := t.Title == tt.Title
	sameTrack = sameTrack && t.Artist == tt.Artist
	sameTrack = sameTrack && t.Album == tt.Album
	sameTrack = sameTrack && t.TrackNumber == tt.TrackNumber
	sameTrack = sameTrack && t.Length == tt.Length

	if !sameTrack {
		return NewTrack
	}

	return t.Stat - tt.Stat
}

func (t *TrackSignal) String() string {
	return fmt.Sprintf("%s\t%s\t%s", t.Artist, t.Album, t.Title)
}

func (t *TrackSignal) RelPath() string {
	return filepath.Join(t.Artist, t.Album, fmt.Sprintf("%d - %s", t.TrackNumber, t.Title))
}
