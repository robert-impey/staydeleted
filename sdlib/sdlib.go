package sdlib

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
)

type ActionForFile struct {
	File, Action string
}

const SdFolderName = ".stay-deleted"

func GetSdFolder(file string) (string, error) {
	dir := filepath.Dir(file)
	attemptedAbsSdFolder := filepath.Join(dir, SdFolderName)
	absSdFolder, err := filepath.Abs(attemptedAbsSdFolder)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find the absolute path of '%v'!",
			attemptedAbsSdFolder)
		return "", err
	} else {
		return absSdFolder, nil
	}
}

func GetSdFile(file string) (string, error) {
	sdFolder, err := GetSdFolder(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get sd folder for '%v'!", file)
		return "", err
	}

	fileBase := filepath.Base(file)
	data := []byte(fileBase)
	return filepath.Join(sdFolder, fmt.Sprintf("%x.txt", md5.Sum(data))), nil
}
