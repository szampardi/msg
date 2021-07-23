// Copyright (c) 2019-2021 SILVANO ZAMPARDI, All rights reserved.
// This source code license can be found in the LICENSE file in the root directory of this source tree.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"

	log "github.com/szampardi/msg"
)

var (
	l          log.Logger                                                                                    //
	data       = make(map[string]interface{})                                                                //
	dataIndex  []string                                                                                      //
	name                                      = flag.String("n", os.Args[0], "set name for verbose logging") //
	logfmt     log.Format                     = log.Formats[log.PlainFormat]                                 //
	loglvl     log.Lvl                        = log.LNotice                                                  //
	logcolor                                  = flag.Bool("c", false, "colorize output")                     ////
	_templates []struct {
		s      string
		isFile bool
	}
	unsafe                *bool     = flag.Bool("u", unsafeMode(), fmt.Sprintf("allow evaluation of dangerous template functions (%v)", unsafeFuncs())) //
	showFns               *bool     = flag.Bool("F", false, "print available template functions and exit")                                              //
	debug                 *bool     = flag.Bool("D", false, "debug init and template rendering activities")                                             //
	startDebuggingOnce    sync.Once                                                                                                                     //
	output                *os.File                                                                                                                      //
	argsfirst             *bool     = flag.Bool("a", false, "output arguments (if any) before stdin (if any), instead of the opposite")                 //
	showVersion           *bool     = flag.Bool("v", false, "print build version/date and exit")                                                        //
	server                *string   = flag.String("s", "", "start a render server on given address")                                                    //
	semver, commit, built           = "v0.0.0-dev", "local", "a while ago"                                                                              //
)

func unsafeFuncs() []string {
	out := []string{}
	for name, info := range templateFnsInfo {
		if info.Unsafe {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func unsafeMode() bool {
	envvar, err := strconv.ParseBool(os.Getenv("XPRINT_UNSAFE"))
	if err != nil {
		return false
	}
	return envvar
}

func logFmts() []string {
	var out []string
	for f := range log.Formats { // lets test them all
		if !strings.Contains(f, "rfc") {
			out = append(out, f)
		}
	}
	sort.Strings(out)
	return out
}

func setFlags() {
	flag.Func(
		"f",
		fmt.Sprintf("logging format (prefix) %v", logFmts()),
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
		`template(s) (string or files). this flag can be specified more than once.
the last template specified in the commandline will be executed,
the others can be accessed with the "template" Action.
`,
		func(value string) error {
			if *debug {
				startDebuggingOnce.Do(usageDebugger)
			}
			_, err := os.Stat(value)
			if err == nil {
				_templates = append(_templates, struct {
					s      string
					isFile bool
				}{value, true})
			} else {
				_templates = append(_templates, struct {
					s      string
					isFile bool
				}{value, false})
			}
			return nil
		},
	)
	flag.Func(
		"o",
		"output to (default is stdout for rendered templates/logs, stderr for everything else)",
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
			output, err = os.OpenFile(value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
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
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("", "  ")
		err := enc.Encode(templateFnsInfo)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
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
	if *server != "" {
		if output == nil {
			output = os.Stderr
		}
		var err error
		l, err = log.New(log.Formats[log.StdFormat].String(), log.Formats[log.DefTimeFmt].String(), loglvl, *logcolor, *name, output)
		if err != nil {
			panic(err)
		}
		u, err := url.Parse(*server)
		if err != nil {
			panic(err)
		}
		proto := strings.Split(u.Scheme, ":")[0]
		var addr string
		if proto != "unix" {
			addr = net.JoinHostPort(u.Hostname(), u.Port())
		} else {
			addr = u.Hostname()
		}
		lis, err := net.Listen(proto, addr)
		if err != nil {
			panic(err)
		}
		l.Noticef("set up %s listener on %s", proto, lis.Addr().String())
		http.HandleFunc("/render", renderServer)
		panic(http.Serve(lis, nil).Error())
	}
	buf := new(bytes.Buffer)
	if len(_templates) > 0 {
		if *debug {
			log.Debugf("%v", _templates)
			startDebuggingOnce.Do(usageDebugger)
		}
		var err error
		tpl := template.New("").Funcs(buildFuncMap(*unsafe))
		files := []string{}
		for n, t := range _templates {
			if !t.isFile {
				tpl, err = tpl.New(fmt.Sprintf("arg%d", n)).Parse(t.s)
				if err != nil {
					panic(err)
				}
			} else {
				files = append(files, t.s)
			}
		}
		if len(files) > 0 {
			tpl.ParseFiles(files...)
		}
		if err := tpl.Execute(buf, data); err != nil {
			panic(err)
		}
		if *debug {
			trackWg.Wait()
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

type jresp struct {
	Status  int         `json:"status"`
	Results interface{} `json:"results,omitempty"`
	Error   error       `json:"error,omitempty"`
}

func renderServer(w http.ResponseWriter, r *http.Request) {
	l.Noticef("new request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
	if *debug {
		b, _ := httputil.DumpRequest(r, true)
		l.Debugf("request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, string(b))
	}
	if r.Method != http.MethodPost {
		l.Errorf("rejected request ( %s %s ) from %s: bad method", r.Method, r.URL, r.RemoteAddr)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var postTemplates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&postTemplates); err != nil {
		l.Errorf("error processing request ( %s %s ) from %s: invalid body: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(jresp{
			Status: http.StatusBadRequest,
			Error: fmt.Errorf(
				`must submit a map[key]value JSON object where key is a string identifier for the template and value is the base64-encoded template itself. a special map["data"] field (also base64 encoded) may be provided, it will not be considered a template but as a data object to apply to the templates in template.Execute`,
			),
		})
		return
	}
	var postedData interface{}
	if v, ok := postTemplates["data"]; ok {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			l.Warningf("error decoding map[data] from request ( %s %s ) from %s: b64dec: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err,
			})
			return
		}
		postedData = string(b)
		delete(postTemplates, "data")
	}
	if len(postTemplates) < 1 {
		l.Infof("request ( %s %s ) from %s: no content", r.Method, r.URL, r.RemoteAddr)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	tpl := template.New("").Funcs(buildFuncMap(*unsafe))
	for n, t := range postTemplates {
		b, err := base64.RawStdEncoding.DecodeString(t)
		if err != nil {
			l.Warningf("error decoding templates from request ( %s %s ) from %s: b64dec: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err,
			})
			return
		}
		tpl, err = tpl.New(n).Parse(string(b))
		if err != nil {
			l.Warningf("error processing request ( %s %s ) from %s: tpl.New: %s", r.Method, r.URL, r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(jresp{
				Status: http.StatusBadRequest,
				Error:  err,
			})
			return
		}
	}
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, postedData); err != nil {
		l.Warningf("error processing request ( %s %s ) from %s: tpl.Execute: %s", r.Method, r.URL, r.RemoteAddr, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(jresp{
			Status: http.StatusInternalServerError,
			Error:  err,
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, buf)
	if err != nil {
		l.Warningf("error sending response to request ( %s %s ) from %s: %s", r.Method, r.URL, r.RemoteAddr, err)
	}
	l.Infof("processed request ( %s %s ) from %s", r.Method, r.URL, r.RemoteAddr)
}
