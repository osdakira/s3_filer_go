package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os/exec"
	"strings"
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

func ConvertReadableString(buf []byte) (string, error) {
	ftype, err := testFileType(buf)
	if err != nil {
		return ftype, err
	}

	switch {
	case strings.Contains(ftype, "ASCII text"):
		return string(buf), nil
	case strings.Contains(ftype, "gzip"):
		return readGzip(buf)
		// case strings.Contains(ftype, "Apache Parquet"):
		// 	return ReadParquet(buf)
	}

	return "Binary file not shown", nil
}

func testFileType(buf []byte) (string, error) {
	cmd := exec.Command("file", "-b", "-")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, string(buf))
	stdin.Close()
	out, err := cmd.Output()
	return string(out), err
}

func readGzip(buf []byte) (string, error) {
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
