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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

	action := "delete"

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
