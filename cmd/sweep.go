// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type ActionForFile struct {
	file, action string
}

// sweepCmd represents the sweep command
var sweepCmd = &cobra.Command{
	Use:   "sweep",
	Short: "Sweep directories of files marked for deletion.",
	Long: `Walk through the directories given in the command line args
looking for files that have been marked for deletion.
`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			err := sweep(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sweepCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sweepCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sweepCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sweep(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return sweepDirectory(path)
	} else {
		return sweepFrom(path)
	}
}

func sweepFrom(sweepFromFileName string) error {
	sweepFromFile, err := os.Open(sweepFromFileName)
	defer sweepFromFile.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open file to sweep from: %v\n", err)
		return err
	}

	input := bufio.NewScanner(sweepFromFile)
	for input.Scan() {
		directoryToSweep := input.Text()
		err := sweepDirectory(directoryToSweep)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to sweep from: '%v' - '%v'\n", directoryToSweep, err)
			continue
		}
	}

	return nil
}

func sweepDirectory(directoryToSweep string) error {
	var absDirectoryToSweep, err = filepath.Abs(directoryToSweep)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find the absolute path for '%v'!\n",
			directoryToSweep)
		return err
	}

	fmt.Printf("Sweeping: '%v'\n", absDirectoryToSweep)
	filesToDelete := make([]string, 0)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return err
		}

		if info.IsDir() && info.Name() == sdFolderName {
			sdFolder := path
			fmt.Printf("Search SD folder '%v'\n", sdFolder)
			containingFolder := filepath.Dir(sdFolder)
			fmt.Printf("Containing folder '%v'\n", containingFolder)

			sdFiles, err := filepath.Glob(filepath.Join(sdFolder, "*.txt"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return err
			}

			for _, sdFile := range sdFiles {
				fmt.Printf("SD File '%v'\n", sdFile)
				actionForFile, err := getActionForFile(sdFile, containingFolder)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
					return err
				}

				if actionForFile.action == "delete" {
					fmt.Printf("Deleting '%v'\n", actionForFile.file)
					filesToDelete = append(filesToDelete, actionForFile.file)
				} else if actionForFile.action == "keep" {
					fmt.Printf("Keeping '%v'\n", actionForFile.file)
				} else {
					fmt.Fprintf(os.Stderr, "Unrecognised action '%v' from '%v'!\n",
						actionForFile.action, sdFile)
				}
			}
		}

		return nil
	}

	err = filepath.Walk(absDirectoryToSweep, walker)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	for _, fileToDelete := range filesToDelete {
		fmt.Printf("Deleting: '%v'\n", fileToDelete)
		os.RemoveAll(fileToDelete)
	}

	return nil
}

func getActionForFile(sdFileName, containingFolder string) (ActionForFile, error) {
	sdFile, err := os.Open(sdFileName)
	defer sdFile.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return ActionForFile{"", ""}, err
	}

	input := bufio.NewScanner(sdFile)
	input.Scan()
	fileToProcessName := filepath.Join(containingFolder, input.Text())
	input.Scan()
	action := input.Text()

	return ActionForFile{fileToProcessName, action}, nil
}
