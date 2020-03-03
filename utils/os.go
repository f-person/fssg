package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Create directory "dir" if it does not exist
// otherwise create after recursively deleting.
func CreateDir(dir string) (err error) {
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0755)
			return
		} else {
			return
		}
	} else {
		err = os.RemoveAll(dir)
		if err != nil {
			return
		}

		err = os.Mkdir(dir, 0755)
		if err != nil {
			return
		}
	}

	return
}

// Copy file src to dst
func CopyFile(src, dst string) (err error) {
	srcStat, err := os.Stat(src)
	if err != nil {
		return
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s: not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	return
}

// Recursively copy directory src to dst
func CopyDir(src, dst string) (err error) {
	err = CreateDir(dst)
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
