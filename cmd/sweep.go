// Copyright © 2018 Robert Impey robert.impey@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/robert-impey/staydeleted/sdlib"
	"github.com/spf13/cobra"
)

var NumRepeats int
var Period int32
var LogsDir string

var ExpiryMonths int
var Verbose bool

var OutWriter io.Writer
var ErrWriter io.Writer

// sweepCmd represents the sweep command
var sweepCmd = &cobra.Command{
	Use:   "sweep",
	Short: "Sweep directories of files marked for deletion.",
	Long: `Walk through the directories given in the command line args
looking for files that have been marked for deletion.
`,
	Run: func(cmd *cobra.Command, args []string) {
		sweep(args)
	},
}

func init() {
	rootCmd.AddCommand(sweepCmd)
	sweepCmd.Flags().IntVarP(&NumRepeats, "repeats", "r", 0,
		"The number of times to repeat the sweeping.")
	sweepCmd.Flags().Int32VarP(&Period, "period", "p", 3600,
		"The number of seconds in the waiting period. A random time during the period is chosen.")
	sweepCmd.Flags().StringVarP(&LogsDir, "logs", "l", "",
		"The logs directory.")
	sweepCmd.Flags().IntVarP(&ExpiryMonths, "expiry", "e", 12,
		"The number of months before SD files expire.")
	sweepCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "Print verbosely.")
}

func sweep(paths []string) {
	if len(LogsDir) > 0 {
		rootLogFolder, err := filepath.Abs(LogsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
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
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		OutWriter = outLogFile

		errLogFile, err := os.Create(errLogFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
		ErrWriter = errLogFile
	} else {
		OutWriter = os.Stdout
		ErrWriter = os.Stderr
	}

	if NumRepeats < 1 {
		sweepPaths(paths)
	} else {
		for i := 0; i < NumRepeats; i++ {
			firstWait := rand.Int31n(Period)
			time.Sleep(time.Duration(firstWait) * time.Second)
			if Verbose {
				fmt.Fprintf(OutWriter, "Run: %d at %s\n", i,
					time.Now().Format("2006-01-02 15:04:05"))
			}

			sweepPaths(paths)

			if i < NumRepeats-1 {
				secondWait := Period - firstWait
				time.Sleep(time.Duration(secondWait) * time.Second)
			}
		}
	}
}

func sweepPaths(paths []string) {
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(ErrWriter, "%v\n", err)
			continue
		}

		if stat.IsDir() {
			err := sweepDirectory(path)
			if err != nil {
				fmt.Fprintf(ErrWriter, "%v\n", err)
			}
		} else {
			err := sweepFrom(path)
			if err != nil {
				fmt.Fprintf(ErrWriter, "%v\n", err)
			}
		}
		if err != nil {
			fmt.Fprintf(ErrWriter, "%v\n", err)
		}
	}
}

func sweepFrom(sweepFromFileName string) error {
	var directoriesToSweepFrom, err = sdlib.ReadSweepFromFile(sweepFromFileName)
	if err != nil {
		_, err := fmt.Fprintf(ErrWriter, "Unable to read file to sweep from '%v' - '%v'\n", sweepFromFileName, err)
		if err != nil {
			return err
		}
	}

	for _, directoryToSweepFrom := range directoriesToSweepFrom {
		err := sweepDirectory(directoryToSweepFrom)
		if err != nil {
			return err
		}
	}

	return nil
}

func sweepDirectory(directoryToSweep string) error {
	type fileToDelete struct {
		Path, SDFile string
	}

	var absDirectoryToSweep, err = filepath.Abs(directoryToSweep)
	if err != nil {
		fmt.Fprintf(ErrWriter, "Unable to find the absolute path for '%v' - '%v'!\n",
			directoryToSweep, err)
		return err
	}

	sdExpiryCutoff := time.Now().AddDate(0, -1*ExpiryMonths, 0)

	re, _ := regexp.Compile(`[0-9a-fA-F]+.txt`)
	if Verbose {
		fmt.Fprintf(OutWriter, "Sweeping: '%v'\n", absDirectoryToSweep)
	}
	filesToDelete := make([]fileToDelete, 0)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(ErrWriter, "%v\n", err)
			return err
		}

		if info.IsDir() && info.Name() == sdlib.SdFolderName {
			sdFolder := path
			if Verbose {
				fmt.Fprintf(OutWriter, "Search SD folder '%v'\n", sdFolder)
			}
			containingFolder := filepath.Dir(sdFolder)
			if Verbose {
				fmt.Fprintf(OutWriter, "Containing folder '%v'\n", containingFolder)
			}

			sdFiles, err := filepath.Glob(filepath.Join(sdFolder, "*.txt"))
			if err != nil {
				fmt.Fprintf(ErrWriter, "%v\n", err)
				return err
			}

			// Remove emptied sd folders
			if len(sdFiles) == 0 {
				fmt.Fprintf(OutWriter, "Adding empty SD folder '%s' to the delete list\n", sdFolder)
				filesToDelete = append(filesToDelete, fileToDelete{Path: sdFolder, SDFile: ""})
			}

			for _, sdFile := range sdFiles {
				sdStat, err := os.Stat(sdFile)
				if err != nil {
					fmt.Fprintf(ErrWriter, "%v\n", err)
					return err
				}

				if !re.Match([]byte(sdStat.Name())) {
					fmt.Fprintf(OutWriter, "'%v' is not a legal name for SD file - deleting.\n",
						sdFile)
					filesToDelete = append(filesToDelete, fileToDelete{sdFile, ""})
					continue
				}

				if sdStat.ModTime().Before(sdExpiryCutoff) {
					fmt.Fprintf(OutWriter, "Adding old SD file '%v' from %s to the delete list\n",
						sdFile,
						sdStat.ModTime().Format("2006-01-02 15:04:05"))
					filesToDelete = append(filesToDelete, fileToDelete{sdFile, ""})
					continue
				}

				if Verbose {
					fmt.Fprintf(OutWriter, "SD File '%v'\n", sdFile)
				}
				actionForFile, err := sdlib.GetActionForFile(sdFile, containingFolder, ErrWriter)
				if err != nil {
					fmt.Fprintf(ErrWriter, "%v\n", err)
					return err
				}

				if actionForFile.Action == "delete" {
					if _, err := os.Stat(actionForFile.File); os.IsNotExist(err) {
						if Verbose {
							fmt.Fprintf(OutWriter, "'%v' already deleted.\n", actionForFile.File)
						}
						continue
					}
					fmt.Fprintf(OutWriter, "Adding '%v' to the delete list\n", actionForFile.File)
					filesToDelete = append(filesToDelete, fileToDelete{actionForFile.File, actionForFile.SdFile})
				} else if actionForFile.Action == "keep" {
					if Verbose {
						fmt.Fprintf(OutWriter, "Keeping '%v'\n", actionForFile.File)
					}
				} else {
					fmt.Fprintf(ErrWriter, "Unrecognised action '%v' from '%v'!\n",
						actionForFile.Action, sdFile)
					fmt.Fprintf(OutWriter, "Adding unreadable SD file '%v' from %s to the delete list\n",
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
		_, err := fmt.Fprintf(ErrWriter, "%v\n", err)
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
		fmt.Fprintf(OutWriter, "%v\n", deleteMessage)

		err = os.RemoveAll(fileToDelete.Path)
		if err != nil {
			fmt.Fprintf(ErrWriter, "%v\n", err)
			if errors.As(err, &pe) {
				fmt.Fprintf(ErrWriter,
					"Failed to remove %v from %v - Removing the SD file\n",
					pe.Path, fileToDelete.SDFile)

				err = os.RemoveAll(fileToDelete.SDFile)
				if err != nil {
					fmt.Fprintf(ErrWriter, "%v\n", err)
				}
			}
		}
	}

	return nil
}
