package main

import (
	"context"
	"github.com/raphaelreyna/BufferKing/internal/app"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		panic("not enough args (rootPath device)")
	}

	if !parec.Available() {
		panic("parec installation not found")
	}

	c := &app.Conf{
		Root:                   os.Args[1],
		Device:                 os.Args[2],
		ObjectPath:             "/org/mpris/MediaPlayer2",
		Format:                 "wav",
		SaveIncompletesSkipped: false,
		SaveIncompletesPaused:  false,
		Color:                  true,
	}

	a := &app.App{
		Conf:       c,
		SignalChan: make(chan *signal.TrackSignal),
	}

	if err := a.LoadConf(); err != nil {
		panic(err)
	}

	if err := a.StartListening(context.TODO()); err != nil {
		panic(err)
	}

	if err := a.Run(context.TODO()); err != nil {
		panic(err)
	}
}
