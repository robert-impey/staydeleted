package sdlib

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
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

func GetActionForFile(sdFileName, containingFolder string, errWriter io.Writer) (ActionForFile, error) {
	sdFile, err := os.Open(sdFileName)
	defer sdFile.Close()

	if err != nil {
		fmt.Fprintf(errWriter, "%v\n", err)
		return ActionForFile{"", ""}, err
	}

	input := bufio.NewScanner(sdFile)
	input.Scan()
	fileToProcessName := filepath.Join(containingFolder, input.Text())
	input.Scan()
	action := input.Text()

	return ActionForFile{fileToProcessName, action}, nil
}
