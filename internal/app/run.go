package app

import (
	"context"
	"github.com/raphaelreyna/BufferKing/internal/parec"
	"github.com/raphaelreyna/BufferKing/internal/signal"
)

func (a *App) Run(ctx context.Context) error {
	l := a.Library
	p := a.Parec

	var err error
	var lastTS *signal.TrackSignal

	finishWJ := func(wj *parec.WriteJob) error {
		if wj != nil {
			if completed, _ := wj.Completed(); completed {
				l.MarkStored(wj.Track)
				err = l.FileMarkStored(wj.Track)
				if err != nil {
					return err
				}
				a.Print(colorGreen, CompletedNewRecording, nil)
			} else {
				if a.Conf.SaveIncompletes {
					l.MarkStored(wj.Track)
					err = l.FileMarkStored(wj.Track)
					if err != nil {
						return err
					}
				}
				a.Print(colorYellow, UnableToCompleteSkip, nil)
			}
		}

		return nil
	}

	for ts := range a.SignalChan {
		if l.Stored(&ts.Track) {
			a.Print(colorCyan, TrackFoundIgnoring, &ts.Track)
			continue
		}

		switch diff := lastTS.Compare(ts); diff {
		case signal.NewTrack:
			oldWJ, err := p.NewWriteJob(context.TODO(), &ts.Track)
			if err != nil {
				return err
			}

			if err := finishWJ(oldWJ); err != nil {
				return err
			}

			a.Print(colorRed, TrackStartedRecording, &ts.Track)
		case signal.Paused:
			if !p.RunningWriteJob() {
				continue
			}

			// Get the current job and stop it
			wj := p.WriteJob()
			err = wj.Stop()
			if err != nil {
				panic(err)
			}

			if err := finishWJ(wj); err != nil {
				return err
			}

			a.Print(colorYellow, UnableToCompletePause, nil)
		case signal.Resumed:
			a.Print(colorYellow, TrackUnableToResume, &ts.Track)
		case signal.None:
			continue
		}

		lastTS = ts
	}

	return err
}
