package main

import (
	"bytes"
    "flag"   
    "fmt"
    "os"
    "./flunky"
)

var Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
}

var server string
var verbose bool
var help bool
var info string
var error string
var get string
var exec string

func init() {
    flag.BoolVar(&help, "h", false, "print usage")
    flag.BoolVar(&verbose, "v", false, "print debug information")
    flag.StringVar(&server, "S", "http://localhost:8081", "server base URL")
	flag.StringVar(&info, "i", "", "log info message")
	flag.StringVar(&error, "e", "", "log error message")
	flag.StringVar(&get, "g", "", "fetch and print endpoint")
	flag.StringVar(&exec, "x", "", "fetch and run endpoint")
}

func main() {
    flag.Parse()
    if help {
        Usage()
        os.Exit(0)
       }

    bs := flunky.NewBuildServer(server, verbose)

	bs.DebugLog(fmt.Sprintf("Server is %s", server))

	if get != "" {
		data, _ := bs.Get(get)
		fmt.Fprintf(os.Stderr, "%s", string(data))
	} else if exec != "" {
		_, _ = bs.Run(exec)
	} else if info != "" {
		buf := bytes.NewBufferString(info)
		_ = bs.Info(buf)
	} else if error != "" {
		buf := bytes.NewBufferString(error)
		_ = bs.Error(buf)
	}
}