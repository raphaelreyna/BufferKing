package main

import (
	"context"
	"fmt"
	"github.com/raphaelreyna/BufferKing/internal/app"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
	flag "github.com/spf13/pflag"
	"os"
	osgn "os/signal"
	"strconv"
)

var Version = ""

func main() {
	// Setup program exit
	retCode := 1
	defer func() { os.Exit(retCode) }()

	// Does the machine have the parec binary for us to use?
	if !parec.Available() {
		fmt.Println("parec installation not found")
		return
	}

	// Setup and parse flags
	var (
		listFormats bool
		listSources bool
		version     bool
	)
	c, p := userConf(&listFormats, &listSources, &version)

	// Does the user just want some basic info or are we recording?
	if version {
		printVersion()
		retCode = 0
		return
	}
	if listFormats {
		if err := printFormats(); err == nil {
			retCode = 0
		}
		return
	}
	if listSources {
		if err := printSources(); err == nil {
			retCode = 0
		}
		return
	}

	// Make sure user gave a valid path to library
	argsCount := len(os.Args)
	if argsCount < 2 {
		fmt.Println("not enough args, need path to root directory for library")
		return
	}
	info, err := os.Stat(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if !info.IsDir() {
		fmt.Println("need path to root directory for library")
		return
	}
	c.Root = os.Args[1]

	// Grab device that will be our audio source to record from
	c.Device, err = source(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create and configure app
	a := &app.App{
		Conf:       c,
		SignalChan: make(chan *signal.TrackSignal),
	}
	if err := a.LoadConf(); err != nil {
		fmt.Println(err)
		return
	}
	a.Listener.Parser = *p

	// Create context, and listen for kill signal
	ctx, cancelFunc := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal)
	osgn.Notify(sigChan, os.Kill)
	go func() {
		<-sigChan
		cancelFunc()
	}()

	// Run bufferking main logic
	if err := a.StartListening(ctx); err != nil {
		fmt.Println(err)
		return
	}

	if err := a.Run(ctx); err != nil {
		fmt.Println(err)
		return
	}

	retCode = 0
}

func source(argv []string) (string, error) {
	var device string
	if len(argv) == 2 {
		sources, err := parec.Sources()
		if err != nil {
			return "", err
		}

		for index, source := range sources {
			fmt.Printf("%d) %s\n", index, source)
		}

		fmt.Print("Record from which source: ")
		var devIndexString string
		fmt.Scanln(&devIndexString)

		devIndex, err := strconv.Atoi(devIndexString)
		if err != nil {
			return "", err
		}

		if devIndex < len(sources) && 0 <= devIndex {
			device = sources[devIndex]
		} else {
			return "", fmt.Errorf("invalid source choice")
		}
	} else {
		device = argv[2]
	}
	return device, nil
}

func printVersion() {
	fmt.Printf("bufferking version: %s\n", Version)
}

func printFormats() error {
	formats, err := parec.Formats()
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, format := range formats {
		fmt.Println(format)
	}

	return nil
}

func printSources() error {
	sources, err := parec.Sources()
	if err != nil {
		fmt.Println(err)
		return err
	}

	for _, source := range sources {
		fmt.Println(source)
	}

	return nil
}

// userConf parses users flag input into a Conf struct
func userConf(formats, sources, version *bool) (*app.Conf, *signal.Parser) {
	c := app.Conf{}
	flag.StringVarP(&c.ObjectPath, "object-path", "o", "/org/mpris/MediaPlayer2", `DBus object path to listen to.`)
	flag.StringVarP(&c.Format, "format", "f", "wav", `Audio format to use when recording.`)
	flag.BoolVarP(&c.SaveIncompletesSkipped, "keep-skipped", "S", false, `Keep incomplete recording due to skipping and mark the track as completed.`)
	flag.BoolVarP(&c.SaveIncompletesPaused, "keep-paused", "P", false, `Keep incomplete recording due to pausing and mark the track as completed.`)
	flag.BoolVarP(&c.RemovePartials, "remove-partials", "r", false, `Remove partial recording parts.`)
	flag.BoolVarP(&c.Color, "color", "c", false, `Use color coded output.`)

	flag.BoolVar(formats, "list-formats", false, `List supported audio formats.`)
	flag.BoolVar(sources, "list-sources", false, `List available audio sources to record.`)
	flag.BoolVarP(version, "version", "v", false, `Print current version.`)

	p := signal.Parser{}
	flag.StringVar(&p.MetaDataKey, "metadata-key", "Metadata", `DBus metadata key`)
	flag.StringVar(&p.TitleKey, "title-key", "xesam:title", `DBus title key`)
	flag.StringVar(&p.ArtistKey, "artist-key", "xesam:artist", `DBus artist key`)
	flag.StringVar(&p.AlbumKey, "album-key", "xesam:album", `DBus album key`)
	flag.StringVar(&p.TrackNumber, "track-no-key", "xesam:trackNumber", `DBus track number key`)
	flag.StringVar(&p.LengthKey, "length-key", "mrpis:length", `DBus track length key`)
	flag.StringVar(&p.StatusKey, "status-key", "PlaybackStatus", `DBus status key`)
	flag.StringVar(&p.PlayToken, "play-token", "Playing", `DBus play token`)
	flag.StringVar(&p.PauseToken, "pause-token", "Paused", `DBus pause token`)

	flag.Parse()

	return &c, &p
}
