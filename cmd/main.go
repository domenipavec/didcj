// Copyright © 2017 Domen Ipavec <domen@ipavec.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/matematik7/didcj/config"
	"github.com/matematik7/didcj/generate"
	"github.com/spf13/cobra"
)

// mainCmd represents the main command
var mainCmd = &cobra.Command{
	Use:   "main",
	Short: "generate main cpp file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("You need to specify filename")
			return
		}
		filename := args[0]
		if !strings.HasSuffix(filename, ".cpp") {
			filename += ".cpp"
		}

		cfg, err := config.Get()
		if err != nil {
			log.Fatal(err)
		}

		getn := "GetN"
		for _, input := range cfg.Input {
			if input.ReturnGenerator == "CONSTANT" && len(input.Inputs) == 0 {
				getn = input.Name
				break
			}
		}

		log.Println("Generating", filename, "...")
		err = generate.MainCpp(filename, getn)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	generateCmd.AddCommand(mainCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mainCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mainCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
