package app

import (
	"context"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
)

func (a *App) Run(ctx context.Context) error {
	l := a.Library
	p := a.Parec
	c := a.Conf

	var err error
	var lastTS *signal.TrackSignal

	for ts := range a.SignalChan {
		//fmt.Println(ts)
		stored := l.Stored(&ts.Track)

		diff := lastTS.Compare(ts)
		//fmt.Println(diff)
		switch diff {
		case signal.NewTrack:
			var finishedWJ *parec.WriteJob
			var printFunc MsgPrinter
			if stored {
				finishedWJ, err = p.StopWriteJob()
				printFunc = a.NewPrinter(colorCyan, TrackFoundIgnoring, &ts.Track)
			} else {
				finishedWJ, err = p.NewWriteJob(context.TODO(), &ts.Track, true)
				printFunc = a.NewPrinter(colorRed, TrackStartedRecording, &ts.Track)
			}
			if err != nil {
				return err
			}

			err := a.finishWJ(finishedWJ, c.SaveIncompletesSkipped, UnableToCompleteSkip)
			if err != nil {
				return err
			}

			printFunc()
		case signal.Paused:
			wj, err := p.StopWriteJob()
			if err != nil {
				return err
			}
			if wj != nil {
				err := a.finishWJ(wj, c.SaveIncompletesPaused, UnableToCompletePause)
				if err != nil {
					return err
				}
			}
		case signal.Resumed:
			if !stored {
				a.Print(colorYellow, TrackUnableToResume, &ts.Track)
			}
		case signal.None:
		}

		lastTS = ts
	}

	return err
}
