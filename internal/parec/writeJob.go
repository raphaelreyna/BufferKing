package parec

import (
	"context"
	"fmt"
	"github.com/raphaelreyna/BufferKing/internal/library"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type WriteJob struct {
	Track *library.Track

	parec *Parec
	cmd   *exec.Cmd

	started time.Time
	stopped time.Time
}

func (wj *WriteJob) Start(ctx context.Context) error {
	p := wj.parec
	wj.started = time.Now()
	if wj.cmd != nil {
		return nil
	}

	p.partsCount += 1

	track := wj.Track
	track.Format = p.Format
	writePath := filepath.Join(p.Root, track.RelPath())

	// Make sure the directory we'll be writing to exists
	dir := filepath.Dir(writePath)
	fileName := wj.FileName()
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	wj.cmd = exec.CommandContext(ctx, "parec",
		"-d", p.Device,
		"--file-format="+p.Format,
		filepath.Join(dir, fileName),
	)

	err = wj.cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (wj *WriteJob) Stop() error {
	cmd := wj.cmd
	if cmd == nil {
		return nil
	}

	err := cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	//Wait for parec to stop
	err = cmd.Wait()
	if err != nil {
		return err
	}

	wj.stopped = time.Now()

	return nil
}

func (wj *WriteJob) Running() bool {
	started := !wj.started.IsZero()
	stopped := !wj.stopped.IsZero()

	return started && !stopped
}

// Completed returns if a track was completely recorded (with some fuzzing ._.)
// The second return value is how long the recording lasted for
func (wj *WriteJob) Completed() (bool, time.Duration) {
	if wj.stopped.IsZero() || wj.started.IsZero() {
		return false, 0
	}

	timeRecorded := wj.stopped.Sub(wj.started)
	dt := timeRecorded - wj.Track.Length

	// If dt >= 0 then at least the entire track was recorded
	// For some reason there is about a 1.5s difference between the recording time and the track length
	// 2.5 seconds should be okay for now ... :(
	return dt >= -2500*time.Millisecond, timeRecorded
}

// FileName gives the filename of the file that this writejob is writing to.
// The file is hidden and includes a part number, formatted as: '.(PART_COUNT)TRACK_NO - TITLE.FORMAT'
func (wj *WriteJob) FileName() string {
	if wj == nil {
		return ""
	}

	t := wj.Track
	t.Format = wj.parec.Format
	return fmt.Sprintf(".(%d)%d - %s.%s", wj.parec.partsCount, t.TrackNumber, t.Title, t.Format)
}
