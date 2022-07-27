/*
Copyright © 2022 Robert Impey robert.impey@hotmail.co.uk

*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/robert-impey/staydeleted/sdlib"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
)

// genPS1Cmd represents the genPS1 command
var genPS1Cmd = &cobra.Command{
	Use:   "genPS1",
	Short: "Generate PS1 Files",
	Long: `Generate a PowerShell script that wraps the command.
One script is generated for each directory in the input file
that is used for sweeping.
`,
	Run: func(cmd *cobra.Command, args []string) {
		genScripts(args)
	},
}

func init() {
	rootCmd.AddCommand(genPS1Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genPS1Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genPS1Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func genScripts(paths []string) {
	directoriesToGenFor := mapset.NewSet[string]()
	for _, path := range paths {
		var directoriesToSweepFrom, err = sdlib.ReadSweepFromFile(path)
		if err != nil {
			_, err := fmt.Fprintf(ErrWriter, "Unable to read file to sweep from '%v' - '%v'\n", path, err)
			if err != nil {
				return
			}
		}

		for _, dir := range directoriesToSweepFrom {
			directoriesToGenFor.Add(dir)
		}
	}

	current, err := user.Current()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	autoGenDir := filepath.Join(current.HomeDir, "autogen", "staydeleted")

	if _, err := os.Stat(autoGenDir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(autoGenDir, os.ModePerm)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
	}

	dirSlice := directoriesToGenFor.ToSlice()
	sort.Strings(dirSlice)
	for _, directoryToGenFor := range dirSlice {
		genScript(directoryToGenFor, autoGenDir)
	}
}

func genScript(dir string, autoGenDir string) {
	fmt.Printf(
		"Generating a PowerShell wrapper script for %s in %s\n",
		dir,
		autoGenDir,
	)

	scriptFileName := strings.ReplaceAll(dir, ":\\", "_")
	scriptFileName = strings.ReplaceAll(scriptFileName, "\\", "_")
	scriptFileName = strings.ReplaceAll(scriptFileName, "/", "_")
	scriptFileName = strings.ReplaceAll(scriptFileName, " ", "_")

	scriptFileName += ".ps1"

	scriptFilePath := filepath.Join(autoGenDir, scriptFileName)

	f, err := os.OpenFile(scriptFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	defer f.Close()

	f.WriteString("# AUTOGEN'D - DO NOT EDIT!\n\n")

	printDate := "date\n\n"

	f.WriteString(printDate)

	f.WriteString("staydeleted.exe sweep -v ")

	f.WriteString(dir)
	f.WriteString("\n\n")

	f.WriteString(printDate)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write script to %v - %v\n", scriptFilePath, err)
	}
}
