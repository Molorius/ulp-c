/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Molorius/ulp-c/pkg/hlp"
	"github.com/spf13/cobra"
)

// hlpCmd represents the hlp command
var hlpCmd = &cobra.Command{
	Use:   "hlp files...",
	Short: "Run the hlp compiler",
	Long: `Compile hlp to an executable binary.

Example:
ulp-c hlp your_code.hlp
This will generate a file out.bin that can be executed by ulp_load_binary().`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		files := make([]hlp.HlpFile, 0)
		for _, a := range args {
			filename := a
			contentBytes, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			hlpFile := hlp.HlpFile{
				Name:     filename,
				Contents: string(contentBytes),
			}
			files = append(files, hlpFile)
		}
		h := hlp.Hlp{}
		err := h.Build(files)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(hlpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hlpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hlpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
