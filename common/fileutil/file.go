package fileutil

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func IsDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}

	panic("not reached")
}
func EnsureFile(filePath string)  {
	dir := filepath.Dir(filePath)
	EnsureDir(dir)
	if !IsFileExists(filePath) {
		os.Create(filePath)
	}
}

func EnsureDir(dir string)  {
	if !IsDirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}
}

func IsFileExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return !fi.IsDir()
	}

	panic("not reached")
}

func IsEmptyDir(directory string) bool {
	if IsFileExists(directory) {
		fds, err := ioutil.ReadDir(directory)
		if err != nil {
			return false
		}
		return len(fds) == 0
	}

	return false
}

// EachChildFile get child fi  and process ,if get error after processing stop, if get a stop flag , stop
func EachChildFile(directory string, process func(path string) (bool, error)) error {
	fds, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, fi := range fds {
		if !fi.IsDir() {
			isContinue, err := process(path.Join(directory, fi.Name()))
			if !isContinue {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func EachDirectory(directory string, process func(path string) (bool, error)) error {
	fds, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, fi := range fds {
		if fi.IsDir() {
			isContinue, err := process(path.Join(directory, fi.Name()))
			if !isContinue {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func AbsolutePath(datadir string, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(datadir, filename)
}
