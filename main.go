package main

import (
	"context"
	"fmt"
	"github.com/raphaelreyna/BufferKing/internal/app"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
	flag "github.com/spf13/pflag"
	"os"
)

func main() {

	c := &app.Conf{}

	listFormats := false

	flag.StringVarP(&c.ObjectPath, "object-path", "o", "/org/mpris/MediaPlayer2", `DBus object path to listen to.`)
	flag.StringVarP(&c.Format, "format", "f", "wav", `Audio format to use when recording.`)
	flag.BoolVarP(&c.SaveIncompletesSkipped, "keep-skipped", "S", false, `Keep incomplete recording due to skipping.`)
	flag.BoolVarP(&c.SaveIncompletesPaused, "keep-paused", "P", false, `Keep incomplete recording due to pausing.`)
	flag.BoolVarP(&c.Color, "color", "c", false, `Use color coded output.`)
	flag.BoolVar(&listFormats, "list-formats", false, `List supported audio formats.`)

	flag.Parse()

	if listFormats {
		formatsHashMap, err := parec.Formats()
		if err != nil {
			panic(err)
		}

		for format, _ := range formatsHashMap {
			fmt.Println(format)
		}

		return
	}

	if len(os.Args) < 3 {
		panic("not enough args (rootPath device)")
	}

	if !parec.Available() {
		panic("parec installation not found")
	}

	c.Root = os.Args[1]
	c.Device = os.Args[2]

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
