package main

import (
	"os"

	"github.com/nexus166/msg"
)

var l *msg.Logger

func init() {
	var err error
	l, err = msg.New(msg.CLIFormat, msg.CLITimeFmt, "msg", true, os.Stdout, msg.LDebug)
	if err != nil {
		panic(err)
	}
}

func main() {
	l.Debugf("%v\n", os.Args)
}
