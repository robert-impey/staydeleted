package cmd

/*
Copyright © 2022 Robert Impey robert.impey@hotmail.co.uk
*/

import (
	"bufio"
	"fmt"
	"github.com/robert-impey/staydeleted/sdlib"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// markFromCmd represents the markFrom command
var markFromCmd = &cobra.Command{
	Use:   "markFrom",
	Short: "Mark all the files in a text file for deletion",
	Long:  `If many files need to be marked for deletion, a text file can be provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			err := markFrom(arg)
			if err != nil {
				fmt.Fprint(os.Stderr, err.Error())
			} else {
				fmt.Println("Success")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(markFromCmd)
}

func markFrom(markFromFileName string) error {
	fmt.Printf("Reading %v\n", markFromFileName)

	markFromFile, err := os.Open(markFromFileName)

	if err != nil {
		return err
	}
	defer markFromFile.Close()

	filesToMark := make([]string, 0)

	input := bufio.NewScanner(markFromFile)
	for input.Scan() {
		fileToMark := input.Text()
		if len(strings.TrimSpace(fileToMark)) == 0 {
			continue
		}
		if strings.HasPrefix(fileToMark, "#") {
			continue
		}

		filesToMark = append(filesToMark, fileToMark)
	}

	action := sdlib.Delete

	for _, fileToMark := range filesToMark {
		err := sdlib.SetActionForFile(fileToMark, action)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		} else {
			fmt.Printf("Marked %v as %v\n", fileToMark, action)
		}
	}

	return nil
}
