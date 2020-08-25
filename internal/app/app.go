package app

import (
	"context"
	"fmt"
	"github.com/raphaelreyna/BufferKing/internal/library"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
	"os"
	"path/filepath"
)

type Conf struct {
	Root                   string
	SaveIncompletesPaused  bool
	SaveIncompletesSkipped bool
	RemovePartials         bool
	// ObjectPath points to the dbus object we're listening to.
	// default: /org/mpris/MediaPlayer2
	ObjectPath string
	// Device can be found using `$ pacmd list | grep .monitor`
	// valid device strings look like: alsa_output.pci-0000_00_1f.3.analog-stereo.monitor
	Device string
	Format string
	Color  bool
}

type App struct {
	Conf       *Conf
	Parec      *parec.Parec
	Library    *library.Library
	Listener   *signal.Listener
	SignalChan chan *signal.TrackSignal
}

// LoadConf expects Conf to not be nil
func (a *App) LoadConf() error {
	var err error
	c := a.Conf
	a.Parec = &parec.Parec{
		Root:   c.Root,
		Device: c.Device,
		Format: c.Format,
	}

	a.Library, err = library.LoadLibrary(c.Root)
	if err != nil {
		return err
	}

	a.Listener = &signal.Listener{
		TrackSignals: a.SignalChan,
		ObjectPath:   c.ObjectPath,
		Parser:       *signal.DefaultParser(),
	}
	return nil
}

func (a *App) StartListening(ctx context.Context) error {
	return a.Listener.Start(ctx)
}

func (a *App) finishWJ(wj *parec.WriteJob, saveIncomplete bool, failMsg string) error {
	l := a.Library
	if wj != nil {
		if completed, _ := wj.Completed(); completed {
			l.Lock()
			l.MarkStored(wj.Track)
			err := l.FileMarkStored(wj.Track, wj.FileName())
			if err != nil {
				l.Unlock()
				return err
			}
			l.Unlock()
			a.Print(colorGreen, CompletedNewRecording, nil)
		} else {
			if saveIncomplete {
				l.Lock()
				l.MarkStored(wj.Track)
				err := l.FileMarkStored(wj.Track, wj.FileName())
				if err != nil {
					l.Unlock()
					return err
				}
				l.Unlock()
			} else if a.Conf.RemovePartials {
				path := filepath.Join(a.Conf.Root,
					wj.Track.Artist,
					wj.Track.Album,
					wj.FileName(),
				)
				if err := os.Remove(path); err != nil {
					return err
				}
			}
			a.Print(colorYellow, failMsg, nil)
		}
	}

	return nil
}

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

const (
	CompletedNewRecording = "completed recording new track"
	UnableToCompleteSkip  = "unable to complete recording track due to early track advancement"
	UnableToCompletePause = "unable to complete recording track due to pause"

	TrackFoundIgnoring    = "track found in library, ignoring:"
	TrackStartedRecording = "started recording new track:"
	TrackUnableToResume   = "unable to resume recording incomplete track due to pause:"
)

type MsgPrinter func()

func (a *App) NewPrinter(color, message string, t *library.Track) MsgPrinter {
	return func() {
		a.Print(color, message, t)
	}
}

func (a *App) Print(color, message string, t *library.Track) {
	var s string
	switch a.Conf.Color {
	case true:
		if t == nil {
			s = fmt.Sprintf("%s%s%s\n\n", color, message, colorReset)
		} else {
			s = fmt.Sprintf("%s%s%s\n%s", color, message, colorReset, t)
		}
	case false:
		if t == nil {
			s = fmt.Sprintf("%s\n\n", message)
		} else {
			s = fmt.Sprintf("%s\n%s", message, t)
		}
	}

	if message == TrackFoundIgnoring {
		s += "\n"
	}

	fmt.Println(s)
}
