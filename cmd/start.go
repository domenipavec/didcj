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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/matematik7/didcj/inventory"
	"github.com/matematik7/didcj/models"
	"github.com/matematik7/didcj/utils"
	"github.com/spf13/cobra"
)

var StartNodes int

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		inv, err := inventory.Init("docker")
		if err != nil {
			log.Fatal(err)
		}
		err = inv.Start(StartNodes)
		if err != nil {
			log.Fatal(err)
		}

		servers, err := inv.Get()
		if err != nil {
			log.Fatal(err)
		}

		executable, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}

		serverJsonFile, err := utils.Json2File("servers", servers)
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(serverJsonFile)

		log.Println("Uploading didcj")
		err = utils.Upload(executable, "didcj", servers...)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Uploading servers.json")
		err = utils.Upload(serverJsonFile, "servers.json", servers...)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Uploading nodeid")
		tmpFile, err := ioutil.TempFile("", "nodeid")
		if err != nil {
			log.Fatal(err)
		}
		for i, server := range servers {
			_, err = tmpFile.WriteAt([]byte(strconv.Itoa(i)), 0)
			if err != nil {
				log.Fatal(err)
			}
			err = utils.Upload(tmpFile.Name(), "nodeid", server)
			if err != nil {
				log.Fatal(err)
			}
		}
		tmpFile.Close()
		os.Remove(tmpFile.Name())

		log.Println("Starting daemon")
		for _, server := range servers {
			go startDaemon(server)
		}

		select {}
	},
}

func startDaemon(server *models.Server) {
	for {
		cmd := exec.Command(
			"sshpass",
			"-p",
			server.Password,
			"ssh",
			"-o",
			"StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", server.Username, server.Ip.String()),
			"./didcj",
			"daemon",
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			log.Println(err)
		}

		log.Printf("Restarting daemon on %s", server.Ip.String())
	}
}

func init() {
	remoteCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	startCmd.Flags().IntVar(&StartNodes, "nodes", 100, "Number of remote nodes")
}
