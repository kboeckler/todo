package main

import "os"

type fileWriter struct {
	filename  string
	appending bool
}

func newFileWriter(filename string, appending bool) *fileWriter {
	writer := fileWriter{filename, appending}
	return &writer
}

func (f *fileWriter) Write(p []byte) (n int, err error) {
	file, err := os.OpenFile(f.filename, f.getFileFlag(), 0644)
	if err != nil {
		return 0, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return file.Write(p)
}

func (f *fileWriter) getFileFlag() int {
	if f.appending {
		return os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		return os.O_CREATE | os.O_WRONLY
	}
}
