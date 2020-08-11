package library

import (
	"context"
	"fmt"
	wtchr "github.com/radovskyb/watcher"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

func (l *Library) Watch(ctx context.Context) error {
	var err error
	if err = ctx.Err(); err != nil {
		return err
	}

	watcher := wtchr.New()
	watcher.FilterOps(wtchr.Remove)

	watcher.AddFilterHook(
		wtchr.RegexFilterHook(
			regexp.MustCompile(`^[0-9]+ \- .+\.[0-9A-Za-z]+$`),
			false,
		),
	)

	re := regexp.MustCompile(`^([0-9]+) \- (.+)\.([0-9A-Za-z]+)$`)

	go func() {
		done := ctx.Done()
		for {
			select {
			case <-done:
				return
			case <-watcher.Closed:
				return
			case event, ok := <-watcher.Event:
				if !ok {
					return
				}

				base := event.Name()
				parts := re.FindStringSubmatch(base)

				if len(parts) != 4 {
					continue
				}

				album := filepath.Dir(event.Path)
				artist := filepath.Dir(album)

				album = filepath.Base(album)
				artist = filepath.Base(artist)

				tn, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Printf("error parsing int in: %+v\n", parts)
				}

				t := &Track{
					Title:       parts[2],
					Album:       album,
					Artist:      artist,
					TrackNumber: int32(tn),
					Format:      parts[3],
				}

				l.Lock()
				l.UnmarkStored(t)
				l.Unlock()
			case err, ok := <-watcher.Error:
				if !ok {
					return
				}
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()

	err = watcher.AddRecursive(l.Root)
	if err != nil {
		return err
	}

	go func() {
		fmt.Printf("started watcher\n")
		if err := watcher.Start(333 * time.Millisecond); err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}
