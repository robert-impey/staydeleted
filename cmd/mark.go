package cmd

// Copyright © 2018 Robert Impey robert.impey@hotmail.co.uk
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

import (
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
			sdlib.SetActionForFile(arg, action)
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
