// Copyright (c) 2019-2021 SILVANO ZAMPARDI, All rights reserved.
// This source code license can be found in the LICENSE file in the root directory of this source tree.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	log "github.com/szampardi/msg"
)

var (
	l                     log.Logger                                                                                                                         //
	data                  = make(map[string]interface{})                                                                                                     //
	dataIndex             []string                                                                                                                           //
	name                                                 = flag.String("n", os.Args[0], "set name for verbose logging")                                      //
	logfmt                log.Format                     = log.Formats[log.PlainFormat]                                                                      //
	loglvl                log.Lvl                        = log.LNotice                                                                                       //
	logcolor                                             = flag.Bool("c", false, "colorize output")                                                          ////
	_template             *template.Template                                                                                                                 //
	unsafe                *bool                          = flag.Bool("u", false, "allow evaluation of dangerous template functions such as cmd,env")         //
	showFns               *bool                          = flag.Bool("F", false, "print available template functions and exit")                              //
	output                *os.File                                                                                                                           //
	argsfirst             *bool                          = flag.Bool("a", false, "output arguments (if any) before stdin (if any), instead of the opposite") //
	showVersion           *bool                          = flag.Bool("v", false, "print build version/date and exit")
	semver, commit, built                                = "v0.0.0-dev", "local", "a while ago"
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
		"t",
		"template (string or file)",
		func(value string) error {
			_, err := os.Stat(value)
			if err == nil {
				_template, err = template.New(value).Funcs(tplFuncMap).ParseFiles(value)
			} else {
				_template, err = template.New(os.Args[0]).Funcs(tplFuncMap).Parse(value)
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
}

func appendData(args []string) {
	for n, s := range args {
		key := fmt.Sprintf("%s%d", "arg", n)
		data[key] = s
		dataIndex = append(dataIndex, key)
	}
}

func init() {
	var err error
	setFlags()
	for !flag.Parsed() {
		flag.Parse()
	}
	if *showVersion {
		fmt.Fprintf(os.Stderr, "github.com/szampardi/msg/cmd/xprint version %s (%s) built %s\n", semver, commit, built)
		os.Exit(0)
	}
	if *showFns {
		fns := []string{}
		for s := range tplFuncMap {
			fns = append(fns, s)
		}
		fmt.Fprintf(os.Stderr, "%v", strings.Join(fns, "\n"))
	}
	if err := log.IsValidLevel(int(loglvl)); err != nil {
		panic(err)
	}
	l, err = log.New(logfmt.String(), log.Formats[log.DefTimeFmt].String(), loglvl, *logcolor, *name, os.Stdout)
	if err != nil {
		panic(err)
	}
	if output != nil {
		l.SetOutput(output)
	}
	args := flag.Args()
	if *argsfirst {
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
			data["stdin"] = string(b)
			dataIndex = append(dataIndex, "stdin")
		}
	}
}

func main() {
	buf := new(bytes.Buffer)
	if _template != nil {
		if err := _template.Execute(buf, data); err != nil {
			panic(err)
		}
	} else {
		for _, s := range dataIndex {
			_, err := fmt.Fprintf(buf, "%s", data[s])
			if err != nil {
				panic(err)
			}
		}
	}
	if buf.Len() < 1 {
		os.Exit(0)
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
}
