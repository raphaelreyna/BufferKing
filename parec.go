package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

type Parec struct {
	Device    string
	Format    string
	WritePath string

	cmd     *exec.Cmd
	running bool
}

func (p *Parec) Start() error {
	if p.cmd != nil {
		return nil
	}

	err := os.MkdirAll(filepath.Dir(p.WritePath), os.ModePerm)
	if err != nil {
		return err
	}

	p.cmd = exec.Command("parec",
		"-d", p.Device,
		"--file-format="+p.Format,
		p.WritePath,
	)

	err = p.cmd.Start()
	if err != nil {
		return err
	}

	p.running = true
	return nil
}

func (p *Parec) Stop() error {
	cmd := p.cmd
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

	p.running = false

	return nil
}

func (p *Parec) Running() bool {
	if p == nil {
		return false
	}
	return p.running
}
