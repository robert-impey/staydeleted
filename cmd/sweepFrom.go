package cmd

// Copyright © 2024 Robert Impey robert.impey@hotmail.co.uk
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

import (
	"fmt"
	"io"
	"os"

	"github.com/robert-impey/staydeleted/sdlib"
	"github.com/spf13/cobra"
)

// sweepFromCmd represents the sweepFrom command
var sweepFromCmd = &cobra.Command{
	Use:   "sweepFrom",
	Short: "Sweep from all the directories listed",
	Long: `The arguments to this command should be text files with
	one directory per line.
	Each directory will be swept.`,
	Run: func(cmd *cobra.Command, args []string) {
		sweepFrom(args)
	},
}

func init() {
	rootCmd.AddCommand(sweepFromCmd)

	sweepFromCmd.Flags().StringVarP(&LogsDir, "logs", "l", "",
		"The logs directory.")
	sweepFromCmd.Flags().IntVarP(&ExpiryMonths, "expiry", "e", 12,
		"The number of months before SD files expire.")
	sweepFromCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "Print verbosely.")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sweepFromCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func sweepFrom(paths []string) {
	outWriter, errWriter, err := sdlib.GetWriters(LogsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	sweepFromPaths(paths, outWriter, errWriter)
}

func sweepFromPaths(paths []string, outWriter io.Writer, errWriter io.Writer) {
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(errWriter, "%v\n", err)
			continue
		}

		if stat.IsDir() {
			fmt.Fprintf(errWriter, "%v\n is a directory!", path)
		} else {
			err := sdlib.SweepFrom(path, ExpiryMonths, outWriter, errWriter, Verbose)
			if err != nil {
				fmt.Fprintf(errWriter, "%v\n", err)
			}
		}
		if err != nil {
			fmt.Fprintf(errWriter, "%v\n", err)
		}
	}
}
