package main

import "os"

type fileWriter struct {
	logfile string
}

func newFileWriter(logfile string) *fileWriter {
	writer := fileWriter{logfile}
	return &writer
}

func (f *fileWriter) Write(p []byte) (n int, err error) {
	file, err := os.OpenFile(f.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return file.Write(p)
}
