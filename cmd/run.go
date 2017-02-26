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
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/matematik7/didcj/utils"
	"github.com/spf13/cobra"
)

const SEND = 0
const RECEIVE = 1
const DEBUG = 2

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "wrap the application for communication and monitoring",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		appFile, err := utils.FindFileBasename("app")
		if err != nil {
			log.Fatal(err)
		}

		var nodeid int
		nodeidFile, err := os.Open("nodeid")
		if err != nil {
			log.Fatal(err)
		}
		_, err = fmt.Fscanf(nodeidFile, "%d", &nodeid)
		nodeidFile.Close()

		appCmd := exec.Command("./" + appFile + ".app")
		stderr, err := appCmd.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}
		stdin, err := appCmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}

		err = appCmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		receiveChannels := make([]chan []byte, 100)
		go func() {
			l, err := net.Listen("tcp", ":3456")
			if err != nil {
				log.Fatal(err)
			}
			defer l.Close()

			for {
				conn, err := l.Accept()
				if err != nil {
					log.Fatal(err)
				}

				go func(c net.Conn) {
					source := parseInt(readOrFatal(c, 4))
					data, err := ioutil.ReadAll(c)
					if err != nil {
						log.Fatal(err)
					}
					receiveChannels[source] <- data
					c.Close()
				}(conn)
			}
		}()

		for {
			buffer := make([]byte, 1)
			n, err := stderr.Read(buffer)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			} else if n == 1 {
				if buffer[0] == RECEIVE {
					source := parseInt(readOrFatal(stderr, 4))
					data := <-receiveChannels[source]
					log.Printf("receive from %d: %v", source, data)
					stdin.Write(formatInt(len(data)))
					stdin.Write(data)
				} else if buffer[0] == SEND {
					target := parseInt(readOrFatal(stderr, 4))
					length := parseInt(readOrFatal(stderr, 4))
					msg := readOrFatal(stderr, length)
					conn, err := net.Dial("tcp", ":3456")
					if err != nil {
						log.Fatal(err)
					}
					conn.Write(formatInt(nodeid))
					conn.Write(msg)
					conn.Close()
					log.Printf("send to %d: %v", target, msg)
				} else if buffer[0] == DEBUG {
					length := parseInt(readOrFatal(stderr, 4))
					log.Print(string(readOrFatal(stderr, length)))
				}
			}
		}

		stdin.Close()
		stderr.Close()

		err = appCmd.Wait()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func readOrFatal(reader io.Reader, n int) []byte {
	buf := make([]byte, n)
	nread, err := reader.Read(buf)
	if err != nil {
		log.Fatal(err)
	} else if nread != n {
		log.Fatal("Too short read")
	}
	return buf
}

func parseInt(data []byte) int {
	if len(data) != 4 {
		log.Fatal("Invalid int data")
	}
	value := 0
	for i, b := range data {
		value |= int(b) << uint(8*i)
	}
	return value
}

func formatInt(value int) []byte {
	data := make([]byte, 4)
	for i := 0; i < 4; i++ {
		data[i] = byte(0xff & (value >> uint(8*i)))
	}
	return data
}
