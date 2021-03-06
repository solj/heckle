package main

import (
	"bytes"
	"encoding/json"
	"flag"
	fnet "flunky/net"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var server string
var verbose bool
var help bool
var info string
var error string
var get string
var exec string

type infoErrorMsg struct {
	Message string
}

func init() {
	flag.BoolVar(&help, "h", false, "print usage")
	flag.BoolVar(&verbose, "v", false, "print debug information")
	flag.StringVar(&server, "S", "http://localhost:8080", "server base URL")
	flag.StringVar(&info, "i", "", "log info message")
	flag.StringVar(&error, "e", "", "log error message")
	flag.StringVar(&get, "g", "", "fetch and print endpoint")
	flag.StringVar(&exec, "x", "", "fetch rendered template and run it.")
}

func parseCmdLine() {
	cmdLineFile, error := os.Open("/proc/cmdline")
        if error != nil {
                fmt.Fprintf(os.Stderr, "%s\n", "ERROR:  Failed to open /proc/cmdline for reading.")
        }

	cmdLineBytes, error := ioutil.ReadAll(cmdLineFile)
        if error != nil {
                fmt.Fprintf(os.Stderr, "ERROR:  Failed to read all from /proc/cmdline.")
        }

	error = cmdLineFile.Close()
        if error != nil {
                fmt.Fprintf(os.Stderr, "ERROR:  Failed to close /proc/cmdline.")
        }

	cmdLineOptions := strings.Split(string(cmdLineBytes), " ")

	for _, value := range cmdLineOptions {
		cmdLineOption := strings.Split(value, "=")

		if cmdLineOption[0] == "flunky" {
			fmt.Fprintf(os.Stdout, "Server = %s.", cmdLineOption[1])
			server = cmdLineOption[1]
		}
	}
}

func main() {
	flag.Parse()
	if help {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	if os.Args[0] == "/opt/bootlocal.sh" {
		parseCmdLine()
		exec = "install"
		bs := fnet.NewBuildServer(server, verbose)
		fmt.Fprintf(os.Stdout, "Getting and executing %s script.\n", exec)
		startTime := time.Now()
		bail := false
		for _, err := bs.Run(exec); err != nil && !bail; _, err = bs.Run(exec) {
			if time.Now().Sub(startTime) >= 300 {
				bail = true
				fmt.Fprintf(os.Stderr, "ERROR: 5 minutes passed and microcore could not get install script.")
			}
			time.Sleep(500000000)
		}
		os.Exit(0)
	} else {

		bs := fnet.NewBuildServer(server, verbose)

		bs.DebugLog(fmt.Sprintf("Server is %s", server))

		if get != "" {
			data, err := bs.Get(get)
			if err != nil {
				os.Exit(255)
			}
			fmt.Fprintf(os.Stdout, "%s", string(data))
		} else if exec != "" {
			fmt.Fprintf(os.Stdout, "Getting and executing %s.", exec)
			status, err := bs.Run(exec)
			if err != nil {
				os.Exit(255)
			}
			os.Exit(status)
		} else if info != "" {
			im := new(infoErrorMsg)
			im.Message = info
			js, _ := json.Marshal(im)
			buf := bytes.NewBufferString(string(js))
			_, err := bs.Post("/info", buf)

			if err != nil {
				os.Exit(255)
			}
		} else if error != "" {
			em := new(infoErrorMsg)
			em.Message = error
			js, _ := json.Marshal(em)
			buf := bytes.NewBufferString(string(js))
			_, err := bs.Post("/error", buf)
			if err != nil {
				os.Exit(255)
			}
		}
	}
}
