package main

import (
	"context"
	"fmt"
	"github.com/raphaelreyna/BufferKing/internal/app"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
	flag "github.com/spf13/pflag"
	"os"
	"strconv"
)

func main() {

	c := &app.Conf{}

	listFormats := false
	listSources := false

	flag.StringVarP(&c.ObjectPath, "object-path", "o", "/org/mpris/MediaPlayer2", `DBus object path to listen to.`)
	flag.StringVarP(&c.Format, "format", "f", "wav", `Audio format to use when recording.`)
	flag.BoolVarP(&c.SaveIncompletesSkipped, "keep-skipped", "S", false, `Keep incomplete recording due to skipping.`)
	flag.BoolVarP(&c.SaveIncompletesPaused, "keep-paused", "P", false, `Keep incomplete recording due to pausing.`)
	flag.BoolVarP(&c.Color, "color", "c", false, `Use color coded output.`)
	flag.BoolVar(&listFormats, "list-formats", false, `List supported audio formats.`)
	flag.BoolVar(&listSources, "list-sources", false, `List available audio sources to record.`)

	flag.Parse()

	if listFormats {
		formats, err := parec.Formats()
		if err != nil {
			panic(err)
		}

		for _, format := range formats {
			fmt.Println(format)
		}

		return
	}
	if listSources {
		sources, err := parec.Sources()
		if err != nil {
			panic(err)
		}

		for _, source := range sources {
			fmt.Println(source)
		}

		return
	}

	argsCount := len(os.Args)
	if argsCount < 2 {
		panic("not enough args, need path to root directory for library")
	}
	info, err := os.Stat(os.Args[1])
	if err != nil {
		panic(err)
	}
	if !info.IsDir() {
		panic("need path to root directory for library")
	}

	c.Root = os.Args[1]

	if argsCount == 2 {
		sources, err := parec.Sources()
		if err != nil {
			panic(err)
		}

		for index, source := range sources {
			fmt.Printf("%d) %s\n", index, source)
		}

		fmt.Print("Record from which source: ")
		var devIndexString string
		fmt.Scanln(&devIndexString)

		devIndex, err := strconv.Atoi(devIndexString)
		if err != nil {
			panic(err)
		}

		if devIndex < len(sources) && 0 <= devIndex {
			c.Device = sources[devIndex]
		}
	} else {
		c.Device = os.Args[2]
	}

	if !parec.Available() {
		panic("parec installation not found")
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
