package signal

import (
	"context"
	"fmt"
	dbus "github.com/godbus/dbus/v5"
	"time"
)

const DefaultDebounce = 500 * time.Millisecond

type Listener struct {
	TrackSignals chan<- *TrackSignal
	ObjectPath   string
	Parser
	DebounceDuration time.Duration

	conn    *dbus.Conn
	sigChan chan *dbus.Signal
}

func (l *Listener) Stop() error {
	if l.conn == nil {
		return nil
	}
	close(l.sigChan)
	close(l.TrackSignals)
	defer l.conn.Close()

	objp := dbus.ObjectPath(l.ObjectPath)
	mopt := dbus.WithMatchObjectPath(objp)
	return l.conn.RemoveMatchSignal(mopt)
}

func (l *Listener) Start(ctx context.Context) error {
	if l.conn != nil {
		return nil
	}
	var err error
	if err = ctx.Err(); err != nil {
		return err
	}

	l.conn, err = dbus.SessionBus()
	if err != nil {
		return err
	}

	objp := dbus.ObjectPath(l.ObjectPath)
	mopt := dbus.WithMatchObjectPath(objp)
	err = l.conn.AddMatchSignal(mopt)
	if err != nil {
		return err
	}

	l.sigChan = make(chan *dbus.Signal, cap(l.TrackSignals))
	l.conn.Signal(l.sigChan)

	if l.DebounceDuration == 0 {
		l.DebounceDuration = DefaultDebounce
	}
	go func() {
		done := ctx.Done()
		dbst := time.Now() // debounce start time
		for {
			select {
			case signal := <-l.sigChan:
				now := time.Now()
				if dt := now.Sub(dbst); dt < l.DebounceDuration {
					continue
				}
				dbst = now

				ts, err := l.Parse(signal)
				if err != nil {
					fmt.Println(err)
					err = l.Stop()
					if err != nil {
						fmt.Println(err)
					}
					break
				}

				l.TrackSignals <- ts
			case <-done:
				err = l.Stop()
				if err != nil {
					fmt.Println(err)
				}
				break
			}
		}
	}()

	return nil
}
