// Copyright © 2018 Robert Impey robert.impey@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robert-impey/staydeleted/sdlib"
	"github.com/spf13/cobra"
)

var Keep bool

// markCmd represents the mark command
var markCmd = &cobra.Command{
	Use:   "mark",
	Short: "Mark a file for deletion or keeping",
	Long: `Files marked for deletion or keeping will be
taken care of by the sweep command.`,
	Run: func(cmd *cobra.Command, args []string) {
		action := "delete"
		if Keep {
			action = "keep"
		}

		for _, arg := range args {
			setActionForFile(arg, action)
		}
	},
}

func init() {
	rootCmd.AddCommand(markCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// markCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	markCmd.Flags().BoolVarP(&Keep, "keep", "k", false, "Keep this file.")
}

func setActionForFile(fileName string, action string) error {
	var absFileName, err = filepath.Abs(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find the absolute path for '%v'!\n", fileName)
		return err
	}

	fmt.Printf("Marking: '%v'!\n", absFileName)
	fileBase := filepath.Base(absFileName)
	sdFileName, err := getSdFile(absFileName)

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

func getSdFile(file string) (string, error) {
	sdFolder, err := sdlib.GetSdFolder(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get sd folder for '%v'!", file)
		return "", err
	}

	fileBase := filepath.Base(file)
	data := []byte(fileBase)
	return fmt.Sprintf("%v/%x.txt", sdFolder, md5.Sum(data)), nil
}
