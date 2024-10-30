package main

import "os"

type fileDeleter struct {
	filename string
}

func newFileDeleter(filename string) *fileDeleter {
	deleter := fileDeleter{filename}
	return &deleter
}

func (f *fileDeleter) Delete() error {
	err := os.Remove(f.filename)
	if err != nil {
		return err
	}

	return nil
}
