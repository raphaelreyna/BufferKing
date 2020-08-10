package parec

import (
	"context"
	"github.com/raphaelreyna/BufferKing/internal/library"
)

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

type Parec struct {
	Root     string
	Device   string
	Format   string
	formats  []string
	writeJob *WriteJob
}

func (p *Parec) WriteJob() *WriteJob {
	return p.writeJob
}

func (p *Parec) RunningWriteJob() bool {
	if p.writeJob == nil {
		return false
	}

	return p.writeJob.Running()
}

func (p *Parec) NewWriteJob(ctx context.Context, t *library.Track) (prevWriteJob *WriteJob, err error) {
	prevWriteJob = p.writeJob

	if p.RunningWriteJob() {
		err = p.writeJob.Stop()
		if err != nil {
			return
		}
	}

	p.writeJob = &WriteJob{
		Track: t,
	}

	err = p.writeJob.Start(ctx, p)
	return
}

func (p *Parec) ValidFormat() (bool, error) {
	if p.formats == nil {
		var err error
		p.formats, err = Formats()
		if err != nil {
			return false, err
		}
	}

	if p.Format == "" {
		return false, nil
	}

	for _, validFormat := range p.formats {
		if validFormat == p.Format {
			return true, nil
		}
	}

	return false, nil
}
