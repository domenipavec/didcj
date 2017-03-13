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
	"os"

	"github.com/matematik7/didcj/compile"
	"github.com/matematik7/didcj/config"
	"github.com/matematik7/didcj/daemon"
	"github.com/matematik7/didcj/inventory"
	"github.com/matematik7/didcj/runner"
	"github.com/matematik7/didcj/utils"
	"github.com/spf13/cobra"
)

var RemoteNodes int

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Get()
		if err != nil {
			log.Fatal(err)
		}

		if RemoteNodes > 0 {
			cfg.NumberOfNodes = RemoteNodes
		}

		inv, err := inventory.Init("docker")
		if err != nil {
			log.Fatal(err)
		}
		servers, err := inv.Get()
		if err != nil {
			log.Fatal(err)
		}

		err = compile.GenerateMessageH(cfg.NumberOfNodes)
		if err != nil {
			log.Fatal(err)
		}

		file, err := utils.FindFileBasename("cpp")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Compiling ...")
		err = compile.Compile(file)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Distributing ...")
		appFile, err := os.Open(file + ".app")
		if err != nil {
			log.Fatal(err)
		}
		url := fmt.Sprintf("/distribute/%s.app/?exec=true", file)
		err = utils.Send(servers[0], url, appFile, nil)
		if err != nil {
			log.Fatal(err)
		}
		appFile.Close()

		log.Println("Running...")
		report := &daemon.RunReport{}
		err = utils.Send(servers[0], "/run/", cfg, report)
		if err != nil {
			log.Fatal(err)
		}

		maxTime := int64(0)
		maxMemory := 0
		for i, report := range report.Reports {
			if report.RunTime > maxTime {
				maxTime = report.RunTime
			}
			if report.MaxMemory > maxMemory {
				maxMemory = report.MaxMemory
			}
			log.Printf(
				"Node %d (ip: %s, msgs: %d, largest: %s, time: %s, memory: %s):",
				i,
				report.Ip,
				report.SendCount,
				utils.FormatSize(report.LargestMsg),
				utils.FormatDuration(report.RunTime),
				utils.FormatSize(report.MaxMemory),
			)
			for _, message := range report.Messages {
				log.Println(message)
			}
		}

		if report.Status == runner.DONE {
			log.Printf("Run successful in %s with %s memory!",
				utils.FormatDuration(maxTime),
				utils.FormatSize(maxMemory),
			)
		} else {
			log.Printf("Run failed in %s with %s memory!",
				utils.FormatDuration(maxTime),
				utils.FormatSize(maxMemory),
			)
		}
	},
}

func init() {
	RootCmd.AddCommand(remoteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// remoteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// remoteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	remoteCmd.Flags().IntVar(&RemoteNodes, "nodes", -1, "Number of remote nodes")
}
