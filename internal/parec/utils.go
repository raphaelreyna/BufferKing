package parec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func Available() bool {
	_, err := exec.LookPath("parec")
	return err == nil
}

func Formats() ([]string, error) {
	formats := []string{}
	cmd := exec.Command("parec", "--list-file-formats")
	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Pull out file extension short form of format from each line
	// this is found in the first column of the output
	buf := bytes.NewBuffer(outputBytes)
	newline := byte('\n')
	for {
		line, err := buf.ReadString(newline)
		eof := err == io.EOF
		if err != nil && !eof {
			return nil, err
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line from parec: %s", line)
		}

		formats = append(formats, parts[0])

		if eof {
			break
		}
	}
	return formats, nil
}
