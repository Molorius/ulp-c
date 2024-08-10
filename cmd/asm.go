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

	"github.com/Molorius/ulp-c/pkg/asm"
	"github.com/spf13/cobra"
)

const flagReservedBytes = "reservedBytes"
const flagOutName = "outName"

// asmCmd represents the asm command
var asmCmd = &cobra.Command{
	Use:   "asm",
	Short: "Run the assembler",
	Long: `Compile ULP assembly to an executable binary. 
Currently this only supports one assembly file.

Example:
ulp-c asm your_code.S
This will generate a file out.bin that can be executed by ulp_load_binary().`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		if len(args) != 1 {
			fmt.Printf("1 assembly file expected but %d found\r\n", len(args))
			os.Exit(1)
		}

		// read the assembly
		filename := args[0]
		contentBytes, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		content := string(contentBytes)

		// compile it
		reservedBytes, _ := cmd.Flags().GetInt(flagReservedBytes)
		assembler := asm.Assembler{}
		bin, err := assembler.BuildFile(content, filename, reservedBytes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// write it to a file
		outputName, _ := cmd.Flags().GetString(flagOutName)
		f, err := os.Create(outputName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		_, err = f.Write(bin)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(asmCmd)

	asmCmd.Flags().IntP(flagReservedBytes, "r", 8176, "number of bytes reserved for the ULP")
	asmCmd.Flags().StringP(flagOutName, "o", "out.bin", "name of the output file")
}
