package common

import (
	"os"
	"path/filepath"
	"strings"
)

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func IsFile(pathname string) bool {
	fileInfo, err := os.Stat(pathname)
	if err != nil {
		return false
	}
	return fileInfo.Mode().IsRegular()
}

func TildePath(path string) (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, userHome) {
		subpath, err := filepath.Rel(userHome, path)
		if err != nil {
			return "", nil
		}
		path = filepath.Join("~", subpath)
	}

	return path, nil
}
