package signal

import (
	"fmt"
	"github.com/raphaelreyna/BufferKing/internal/library"
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

func (s Status) String() string {
	switch s {
	case Play:
		return "Play"
	case Pause:
		return "Pause"
	case None:
		return ""
	case Paused:
		return "Paused Playing"
	case Resumed:
		return "Resumed Playing"
	case NewTrack:
		return "New Track"
	}

	return "INVALID_STATUS"
}

type TrackSignal struct {
	library.Track
	Status
}

// Compare compares the old tracksignal to the new tracksignal
func (t *TrackSignal) Compare(tt *TrackSignal) Status {
	if t == nil && tt != nil {
		if tt.Status == Play {
			return NewTrack
		}
		return Paused
	}

	sameTrack := t.Title == tt.Title
	sameTrack = sameTrack && t.Artist == tt.Artist
	sameTrack = sameTrack && t.Album == tt.Album
	sameTrack = sameTrack && t.TrackNumber == tt.TrackNumber
	sameTrack = sameTrack && t.Length == tt.Length
	if !sameTrack {
		return NewTrack
	}

	return t.Status - tt.Status
}

func (t *TrackSignal) String() string {
	return fmt.Sprintf("%s - %s", t.Status, t.Track)
}
