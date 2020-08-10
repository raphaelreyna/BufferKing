package signal

import (
	dbus "github.com/godbus/dbus/v5"
	"github.com/raphaelreyna/BufferKing/internal/library"
	"time"
)

type Parser struct {
	MetaDataKey string

	TitleKey    string
	ArtistKey   string
	AlbumKey    string
	TrackNumber string
	LengthKey   string
	LengthUnit  time.Duration

	StatusKey  string
	PlayToken  string
	PauseToken string
}

func DefaultParser() *Parser {
	return &Parser{
		MetaDataKey: "Metadata",
		TitleKey:    "xesam:title",
		ArtistKey:   "xesam:artist",
		AlbumKey:    "xesam:album",
		TrackNumber: "xesam:trackNumber",
		LengthKey:   "mpris:length",
		LengthUnit:  time.Microsecond,
		StatusKey:   "PlaybackStatus",
		PlayToken:   "Playing",
		PauseToken:  "Paused",
	}
}

func (p *Parser) Parse(sign *dbus.Signal) (*TrackSignal, error) {
	resp := map[string]dbus.Variant{}
	err := dbus.Store(sign.Body[1:2], &resp)
	if err != nil {
		return nil, err
	}

	var stat Status
	switch sstat := resp[p.StatusKey].Value().(string); sstat {
	case p.PlayToken:
		stat = Play
	case p.PauseToken:
		stat = Pause
	}

	md := resp[p.MetaDataKey].Value().(map[string]dbus.Variant)

	return &TrackSignal{
		Track: library.Track{
			Title:       md[p.TitleKey].Value().(string),
			Artist:      md[p.ArtistKey].Value().([]string)[0],
			Album:       md[p.AlbumKey].Value().(string),
			TrackNumber: md[p.TrackNumber].Value().(int32),
			Length:      time.Duration(md[p.LengthKey].Value().(uint64)) * p.LengthUnit,
		},
		Status: stat,
	}, nil
}
