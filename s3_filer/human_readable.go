package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os/exec"
)

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func GuessFileType(buf []byte) (string, error) {
	cmd := exec.Command("file", "-b", "-")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, string(buf))
	stdin.Close()
	out, err := cmd.Output()
	return string(out), err
}

func ReadGzip(buf []byte) (string, error) {
	r := bytes.NewReader(buf)
	reader, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	out, err := io.ReadAll(reader)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			return "", err
		}
	}
	return string(out), nil
}
