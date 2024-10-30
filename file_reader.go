package main

import (
	"bytes"
	"os"
)

type fileReader struct {
	filename string
}

func newFileReader(filename string) *fileReader {
	writer := fileReader{filename}
	return &writer
}

func (f *fileReader) ReadString() (str string, err error) {
	file, err := os.Open(f.filename)
	if err != nil {
		return "", err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	var buf = bytes.Buffer{}
	_, err = buf.ReadFrom(file)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
