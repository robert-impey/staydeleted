package sdlib

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ActionForFile struct {
	SdFile, File, Action string
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
		return ActionForFile{"", "", ""}, err
	}

	input := bufio.NewScanner(sdFile)
	input.Scan()
	fileToProcessName := filepath.Join(containingFolder, input.Text())
	input.Scan()
	action := input.Text()

	return ActionForFile{sdFileName, fileToProcessName, action}, nil
}

func SetActionForFile(fileName string, action string) error {
	var absFileName, err = filepath.Abs(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find the absolute path for '%v'!\n", fileName)
		return err
	}

	fmt.Printf("Marking: '%v'!\n", absFileName)
	fileBase := filepath.Base(absFileName)
	sdFileName, err := GetSdFile(absFileName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get sd file name for '%v'!",
			absFileName)
		return err
	}

	fmt.Printf("SD File: '%v'!\n", sdFileName)
	sdFolder := filepath.Dir(sdFileName)

	if _, err := os.Stat(sdFolder); os.IsNotExist(err) {
		fmt.Printf("Making directory '%v'\n", sdFolder)
		os.Mkdir(sdFolder, 0755)
	}

	sdFile, err := os.Create(sdFileName)
	defer sdFile.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't create file '%v'!\n",
			sdFileName)
		return err
	}

	fmt.Fprintf(sdFile, "%v\n%v\n", fileBase, action)

	return nil
}

func ReadSweepFromFile(sweepFromFileName string) ([]string, error) {
	sweepFromFile, err := os.Open(sweepFromFileName)

	if err != nil {
		return nil, err
	}
	defer sweepFromFile.Close()

	directoriesToSweep := make([]string, 0)

	input := bufio.NewScanner(sweepFromFile)
	for input.Scan() {
		directoryToSweep := input.Text()
		if len(strings.TrimSpace(directoryToSweep)) == 0 {
			continue
		}
		if strings.HasPrefix(directoryToSweep, "#") {
			continue
		}

		directoriesToSweep = append(directoriesToSweep, directoryToSweep)
	}

	return directoriesToSweep, nil
}
