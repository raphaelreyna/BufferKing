package parec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

func Available() bool {
	_, err := exec.LookPath("parec")
	return err == nil
}

func Formats() ([]string, error) {
	formats := map[string]struct{}{}
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
		if eof {
			break
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line from parec: %s", line)
		}

		formats[parts[0]] = struct{}{}
	}

	formatsSlice := []string{}
	for format, _ := range formats {
		formatsSlice = append(formatsSlice, format)
	}
	return formatsSlice, nil
}

func Sources() ([]string, error) {
	re := regexp.MustCompile(`<(.+\.monitor)`)

	cmd := exec.Command("pacmd", "list")
	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	hashmap := map[string]struct{}{}
	for _, rawSource := range re.FindAllSubmatch(outputBytes, -1) {
		// rawSource is of type [][]byte where the first element is the entire matching string
		if len(rawSource) != 2 {
			continue
		}
		hashmap[string(rawSource[1])] = struct{}{}
	}

	sources := []string{}
	for source, _ := range hashmap {
		sources = append(sources, source)
	}

	return sources, nil
}
