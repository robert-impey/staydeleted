package sdlib

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Action int

const (
	NoAction Action = iota
	Delete
	Keep
)

type ActionForFile struct {
	SdFile, File string
	Action       Action
}

const SdFolderName = ".stay-deleted"

func GetActionForBool(keep bool) Action {
	if keep {
		return Keep
	}
	return Delete
}

func getStringForAction(action Action) string {
	if action == Keep {
		return "keep"
	}

	return "delete"
}

func getActionForString(actStr string) (Action, error) {
	if actStr == "delete" {
		return Delete, nil
	} else if actStr == "keep" {
		return Keep, nil
	}

	return NoAction, fmt.Errorf("unable to convert %s to an action", actStr)
}

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
		return ActionForFile{"", "", NoAction}, err
	}

	input := bufio.NewScanner(sdFile)
	input.Scan()
	fileToProcessName := filepath.Join(containingFolder, input.Text())
	input.Scan()
	action, err := getActionForString(input.Text())
	if err != nil {
		fmt.Fprintf(errWriter, "%v\n", err)
		return ActionForFile{"", "", NoAction}, err
	}

	return ActionForFile{sdFileName, fileToProcessName, action}, nil
}

func SetActionForFile(fileName string, action Action) error {
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

	fmt.Fprintf(sdFile, "%v\n%s\n", fileBase, getStringForAction(action))

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

func SweepFrom(sweepFromFileName string, expiryMonths int, outWriter io.Writer, errWriter io.Writer, verbose bool) error {
	var directoriesToSweepFrom, err = ReadSweepFromFile(sweepFromFileName)
	if err != nil {
		_, err := fmt.Fprintf(errWriter, "Unable to read file to sweep from '%v' - '%v'\n", sweepFromFileName, err)
		if err != nil {
			return err
		}
	}

	for _, directoryToSweepFrom := range directoriesToSweepFrom {
		err := SweepDirectory(directoryToSweepFrom, expiryMonths, outWriter, errWriter, verbose)
		if err != nil {
			return err
		}
	}

	return nil
}

func SweepDirectory(directoryToSweep string, expiryMonths int, outWriter io.Writer, errWriter io.Writer, verbose bool) error {
	stat, err := os.Stat(directoryToSweep)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", directoryToSweep)
	}

	type fileToDelete struct {
		Path, SDFile string
	}

	absDirectoryToSweep, err := filepath.Abs(directoryToSweep)
	if err != nil {
		fmt.Fprintf(errWriter, "Unable to find the absolute path for '%v' - '%v'!\n",
			directoryToSweep, err)
		return err
	}

	sdExpiryCutoff := time.Now().AddDate(0, -1*expiryMonths, 0)

	re, _ := regexp.Compile(`[0-9a-fA-F]+.txt`)
	if verbose {
		fmt.Fprintf(outWriter, "Sweeping: '%v'\n", absDirectoryToSweep)
	}
	filesToDelete := make([]fileToDelete, 0)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(errWriter, "%v\n", err)
			return err
		}

		if info.IsDir() && info.Name() == SdFolderName {
			sdFolder := path
			if verbose {
				fmt.Fprintf(outWriter, "Search SD folder '%v'\n", sdFolder)
			}
			containingFolder := filepath.Dir(sdFolder)
			if verbose {
				fmt.Fprintf(outWriter, "Containing folder '%v'\n", containingFolder)
			}

			sdFiles, err := filepath.Glob(filepath.Join(sdFolder, "*.txt"))
			if err != nil {
				fmt.Fprintf(errWriter, "%v\n", err)
				return err
			}

			// Remove emptied sd folders
			if len(sdFiles) == 0 {
				fmt.Fprintf(outWriter, "Adding empty SD folder '%s' to the delete list\n", sdFolder)
				filesToDelete = append(filesToDelete, fileToDelete{Path: sdFolder, SDFile: ""})
			}

			for _, sdFile := range sdFiles {
				sdStat, err := os.Stat(sdFile)
				if err != nil {
					fmt.Fprintf(errWriter, "%v\n", err)
					return err
				}

				if !re.Match([]byte(sdStat.Name())) {
					fmt.Fprintf(outWriter, "'%v' is not a legal name for SD file - deleting.\n",
						sdFile)
					filesToDelete = append(filesToDelete, fileToDelete{sdFile, ""})
					continue
				}

				if sdStat.ModTime().Before(sdExpiryCutoff) {
					fmt.Fprintf(outWriter, "Adding old SD file '%v' from %s to the delete list\n",
						sdFile,
						sdStat.ModTime().Format("2006-01-02 15:04:05"))
					filesToDelete = append(filesToDelete, fileToDelete{sdFile, ""})
					continue
				}

				if verbose {
					fmt.Fprintf(outWriter, "SD File '%v'\n", sdFile)
				}
				actionForFile, err := GetActionForFile(sdFile, containingFolder, errWriter)
				if err != nil {
					fmt.Fprintf(errWriter, "%v\n", err)
					return err
				}

				if actionForFile.Action == Delete {
					if _, err := os.Stat(actionForFile.File); os.IsNotExist(err) {
						if verbose {
							fmt.Fprintf(outWriter, "'%v' already deleted.\n", actionForFile.File)
						}
						continue
					}
					fmt.Fprintf(outWriter, "Adding '%v' to the delete list\n", actionForFile.File)
					filesToDelete = append(filesToDelete, fileToDelete{actionForFile.File, actionForFile.SdFile})
				} else if actionForFile.Action == Keep {
					if verbose {
						fmt.Fprintf(outWriter, "Keeping '%v'\n", actionForFile.File)
					}
				} else {
					fmt.Fprintf(errWriter, "Unrecognised action '%v' from '%v'!\n",
						actionForFile.Action, sdFile)
					fmt.Fprintf(outWriter, "Adding unreadable SD file '%v' from %s to the delete list\n",
						sdFile,
						sdStat.ModTime().Format("2006-01-02 15:04:05"))
					filesToDelete = append(filesToDelete, fileToDelete{sdFile, ""})
				}
			}
		}

		return nil
	}

	err = filepath.Walk(absDirectoryToSweep, walker)
	if err != nil {
		_, err := fmt.Fprintf(errWriter, "%v\n", err)
		if err != nil {
			return err
		}
		return err
	}

	var pe *fs.PathError
	for _, fileToDelete := range filesToDelete {
		var deleteMessage = fmt.Sprintf("Deleting '%v'", fileToDelete.Path)

		if len(fileToDelete.SDFile) > 0 {
			deleteMessage += fmt.Sprintf(" as instructed by '%v'", fileToDelete.SDFile)
		}
		fmt.Fprintf(outWriter, "%v\n", deleteMessage)

		err = os.RemoveAll(fileToDelete.Path)
		if err != nil {
			fmt.Fprintf(errWriter, "%v\n", err)
			if errors.As(err, &pe) {
				fmt.Fprintf(errWriter, "Failed to remove %v from %v\n", pe.Path, fileToDelete.SDFile)

				if err != nil {
					fmt.Fprintf(errWriter, "%v\n", err)
				}
			}
		}
	}

	return nil
}

func GetWriters(logsDir string) (io.Writer, io.Writer, error) {
	if len(logsDir) > 0 {
		rootLogFolder, err := filepath.Abs(logsDir)
		if err != nil {
			return nil, nil, err
		}

		if _, err := os.Stat(rootLogFolder); os.IsNotExist(err) {
			fmt.Printf("Making root logs directory '%v'\n", rootLogFolder)
			os.Mkdir(rootLogFolder, 0755)
		}

		sdLogFolder := filepath.Join(rootLogFolder, "staydeleted")
		if _, err := os.Stat(sdLogFolder); os.IsNotExist(err) {
			fmt.Printf("Making staydeleted logs directory '%v'\n", sdLogFolder)
			os.Mkdir(sdLogFolder, 0755)
		}

		timeStr := time.Now().Format("2006-01-02_15.04.05")
		outLogFileName := filepath.Join(sdLogFolder, fmt.Sprintf("%s.log", timeStr))
		errLogFileName := filepath.Join(sdLogFolder, fmt.Sprintf("%s.err", timeStr))

		outLogFile, err := os.Create(outLogFileName)
		if err != nil {
			return nil, nil, err
		}

		errLogFile, err := os.Create(errLogFileName)
		if err != nil {
			return nil, nil, err
		}

		return outLogFile, errLogFile, nil
	} else {
		return os.Stdout, os.Stderr, nil
	}
}
