package main

import (
	"fmt"
	dbus "github.com/godbus/dbus/v5"
	"os"
	"path/filepath"
	"time"
)

// device can be found using `$ pacmd list | grep .monitor`
// valid device strings look like: alsa_output.pci-0000_00_1f.3.analog-stereo.monitor
const objPath = "/org/mpris/MediaPlayer2"

const (
	colorReset = "\033[0m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func main() {
	if len(os.Args) < 3 {
		panic("not enough args (rootPath device)")
	}
	root := os.Args[1]
	device := os.Args[2]

	l, err := LoadLibrary(root)
	if err != nil {
		panic(err)
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = conn.AddMatchSignal(dbus.WithMatchObjectPath(objPath))
	if err != nil {
		panic(err)
	}

	sigChan := make(chan *dbus.Signal, 1)
	conn.Signal(sigChan)

	var debounce bool
	var currTrack *TrackSignal
	var parec *Parec
	for sig := range sigChan {
		// Filter out repeated messages from dbus (idk why this is happening)
		if debounce {
			continue
		}
		go func() {
			debounce = true
			time.Sleep(500 * time.Millisecond)
			debounce = false
		}()

		track, err := ParseSignal(sig)
		if err != nil {
			panic(err)
		}

		switch diff := currTrack.Compare(track); diff {
		case NewTrack:
			if parec.Running() {
				err = parec.Stop()
				if err != nil {
					panic(err)
				}
				l.MarkStored(currTrack)
				fmt.Printf("%sfinished recording%s\n\n", colorGreen, colorReset)
			}

			currTrack = track

			if l.Stored(track) {
				fmt.Printf("%strack found in library, not recording:%s\n%s\n\n", colorCyan, colorReset, track)
				continue
			}
			parec = &Parec{
				Device:    device,
				Format:    "wav",
				WritePath: filepath.Join(root, track.RelPath()+".wav"),
			}

			err = parec.Start()
			if err != nil {
				panic(err)
			}

			fmt.Printf("%sstarted recording new track:%s\n%s\n\n", colorRed, colorReset, track)
			continue
		case Pause:
			fallthrough
		case Play:
			fallthrough
		case Paused:
			fallthrough
		case Resumed:
			fallthrough
		case None:
			continue
		}
	}
}
