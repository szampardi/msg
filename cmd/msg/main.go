// Copyright (c) 2019 SILVANO ZAMPARDI, All rights reserved.
// This source code license can be found in the LICENSE file in the root directory of this source tree.

// WIP ---- flag parsing fmt instructions is broken

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	log "github.com/szampardi/msg"
)

var (
	l         log.Logger                                   //
	data      []interface{}                                //
	logfmt    log.Format    = log.Formats[log.PlainFormat] //
	loglvl    log.Lvl       = log.LInfo                    //
	logcolor  bool          = false                        //
	printfmt  string        = "%s"                         //
	output    *os.File                                     //
	argsfirst bool          = false                        //
)

func setFlags() {
	flag.Func(
		"F",
		"logging format (prefix)",
		func(value string) error {
			if v, ok := log.Formats[value]; ok {
				logfmt = v
				return nil
			}
			return fmt.Errorf("invalid format [%s] specified", value)
		},
	)
	flag.Func(
		"l",
		"log level",
		func(value string) error {
			i, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			loglvl = log.Lvl(i)
			return log.IsValidLevel(i)
		},
	)
	flag.Func(
		"c",
		"colorize output",
		func(value string) error {
			b, err := strconv.ParseBool(value)
			if err == nil {
				logcolor = b
			}
			return err
		},
	)
	flag.Func(
		"o",
		"output to",
		func(value string) error {
			switch value {
			case "", "1", "stdout", "/dev/stdout", os.Stdout.Name():
				output = os.Stdout
				return nil
			case "2", "stderr", "/dev/stderr", os.Stderr.Name():
				output = os.Stderr
				return nil
			}
			var err error
			output, err = os.OpenFile(value, os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
			return err
		},
	)
	flag.Func(
		"A",
		"output arguments (if any) before stdin (if any), instead of the opposite",
		func(value string) error {
			b, err := strconv.ParseBool(value)
			if err == nil {
				argsfirst = b
			}
			return err
		},
	)
}

func appendData(args []string) {
	for _, s := range args {
		data = append(data, s)
	}
}

func init() {
	setFlags()
	for !flag.Parsed() {
		flag.Parse()
	}
	flag.Func(
		"f",
		"fmt",
		func(value string) error {
			printfmt = strings.ReplaceAll(strings.ReplaceAll(value, "!", "%"), "|", `\`)
			return nil
		},
	)
	args := flag.Args()
	if len(args) > 0 {
		printfmt = args[0]
		args = args[1:]
		if argsfirst {
			appendData(args)
		} else {
			defer appendData(args)
		}
	}
	var err error
	l, err = log.New(logfmt.String(), log.Formats[log.DefTimeFmt].String(), loglvl, logcolor)
	if err != nil {
		panic(err)
	}
	if output != nil {
		l.SetOutput(output)
	}
	stdin, err := os.Stdin.Stat()
	if err == nil && (stdin.Mode()&os.ModeCharDevice) == 0 {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			l.Errorf("reading %s: %s", stdin.Name(), err)
		} else {
			data = append(data, string(b))
		}
	}
}

func main() {
	switch loglvl {
	case log.LCrit:
		l.Criticalf(printfmt, data...)
	case log.LErr:
		l.Errorf(printfmt, data...)
	case log.LWarn:
		l.Warningf(printfmt, data...)
	case log.LNotice:
		l.Noticef(printfmt, data...)
	case log.LInfo:
		l.Infof(printfmt, data...)
	case log.LDebug:
		l.Debugf(printfmt, data...)
	default:
		panic(log.IsValidLevel(int(loglvl)))
	}
}
