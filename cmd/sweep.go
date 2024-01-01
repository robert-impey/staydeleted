package cmd

// Copyright © 2018 Robert Impey robert.impey@hotmail.co.uk
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
	"github.com/robert-impey/staydeleted/sdlib"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

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
			err := sdlib.SweepDirectory(path, ExpiryMonths, OutWriter, ErrWriter, Verbose)
			if err != nil {
				fmt.Fprintf(ErrWriter, "%v\n", err)
			}
		} else {
			err := sdlib.SweepFrom(path, ExpiryMonths, OutWriter, ErrWriter, Verbose)
			if err != nil {
				fmt.Fprintf(ErrWriter, "%v\n", err)
			}
		}
		if err != nil {
			fmt.Fprintf(ErrWriter, "%v\n", err)
		}
	}
}
