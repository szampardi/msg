// Copyright (c) 2019 SILVANO ZAMPARDI, All rights reserved.
// This source code license can be found in the LICENSE file in the root directory of this source tree.

// WIP ---- flag parsing fmt instructions is broken

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"text/template"

	log "github.com/szampardi/msg"
)

var (
	l         log.Logger         //
	data      []interface{}      //
	name                         = flag.String("n", os.Args[0], "set name for verbose logging")
	logfmt    log.Format         = log.Formats[log.PlainFormat] //
	loglvl    log.Lvl            = log.LInfo                    //
	logcolor  bool               = false                        //
	_template *template.Template                                //
	output    *os.File                                          //
	argsfirst bool               = false                        //
)

func setFlags() {
	flag.Func(
		"f",
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
		"t",
		"template (string or file)",
		func(value string) error {
			var err error
			_template, err = template.ParseFiles(value)
			if err != nil {
				_template, err = template.New(os.Args[0]).Parse(value)
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
		"a",
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
	var err error
	setFlags()
	for !flag.Parsed() {
		flag.Parse()
	}
	if err := log.IsValidLevel(int(loglvl)); err != nil {
		panic(err)
	}
	l, err = log.New(logfmt.String(), log.Formats[log.DefTimeFmt].String(), loglvl, logcolor, *name)
	if err != nil {
		panic(err)
	}
	if output != nil {
		l.SetOutput(output)
	}
	args := flag.Args()
	if argsfirst {
		appendData(args)
	} else {
		defer appendData(args)
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
	if len(data) < 1 {
		os.Exit(0)
	}
	if _template != nil {
		buf := new(bytes.Buffer)
		if err := _template.Execute(buf, data); err != nil {
			panic(err)
		}
		switch loglvl {
		case log.LCrit:
			l.Criticalf("%s", buf.String())
		case log.LErr:
			l.Errorf("%s", buf.String())
		case log.LWarn:
			l.Warningf("%s", buf.String())
		case log.LNotice:
			l.Noticef("%s", buf.String())
		case log.LInfo:
			l.Infof("%s", buf.String())
		case log.LDebug:
			l.Debugf("%s", buf.String())
		}
	} else {
		switch loglvl {
		case log.LCrit:
			l.Criticalf("%s", data...)
		case log.LErr:
			l.Errorf("%s", data...)
		case log.LWarn:
			l.Warningf("%s", data...)
		case log.LNotice:
			l.Noticef("%s", data...)
		case log.LInfo:
			l.Infof("%s", data...)
		case log.LDebug:
			l.Debugf("%s", data...)
		}
	}
}
