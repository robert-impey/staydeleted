/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/robert-impey/staydeleted/sdlib"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cobra"
)

// genPS1Cmd represents the genPS1 command
var genPS1Cmd = &cobra.Command{
	Use:   "genPS1",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

	scriptContents := "# AUTOGEN'D - DO NOT EDIT!\n\n"

	printDate := "date\n\n"

	scriptContents += printDate

	scriptContents += "staydeleted.exe sweep "

	scriptContents += dir
	scriptContents += "\n\n"

	scriptContents += printDate

	scriptFileName := strings.ReplaceAll(dir, ":\\", "_")
	scriptFileName = strings.ReplaceAll(scriptFileName, "\\", "_")
	scriptFileName = strings.ReplaceAll(scriptFileName, "/", "_")
	scriptFileName = strings.ReplaceAll(scriptFileName, " ", "_")

	scriptFileName += ".ps1"

	scriptFilePath := filepath.Join(autoGenDir, scriptFileName)

	err := ioutil.WriteFile(scriptFilePath, []byte(scriptContents), 0x755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write script to %v - %v\n", scriptFilePath, err)
	}
}
