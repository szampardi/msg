// COPYRIGHT (c) 2019-2021 SILVANO ZAMPARDI, ALL RIGHTS RESERVED.
// The license for these sources can be found in the LICENSE file in the root directory of this source tree.

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/szampardi/msg/ansi"
)

// Worker class, Worker is a log object used to log messages and Color specifies
// if colored output is to be produced
type worker struct {
	Minion     *log.Logger
	Color      bool
	spFormat   string
	format     string
	timeFormat string
	level      Lvl
}

//NewWorker  Returns an instance of worker class, prefix is the string attached to every log,
// flag determine the log params, color parameters verifies whether we need colored outputs or not
func newWorker(prefix string, format, timeformat string, flag int, color bool, out io.Writer, lvl Lvl) *worker {
	var spFormat string
	if format == "" {
		format = Formats[PlainFormat]._String
	}
	switch format {
	case "yaml":
		spFormat = "yaml"
	case "json":
		spFormat = "json"
	}
	if v, ok := Formats[format]; ok {
		format = v._String
	}
	if timeformat == "" {
		timeformat = time.RFC3339
	}
	if out == nil {
		out = os.Stdout
	}
	return &worker{
		Minion:     log.New(out, prefix, flag),
		Color:      color,
		spFormat:   spFormat,
		format:     format,
		timeFormat: timeformat,
		level:      lvl,
	}
}

// New Returns a new instance of logger class, module is the specific module for which we are logging
// , color defines whether the output is to be colored or not, out is instance of type io.Writer defaults
// to os.Stderr
func New(format, timeformat string, args ...interface{}) (Logger, error) {
	var module string = "msg"
	var color bool = true
	var out io.Writer = os.Stderr
	var level Lvl = LDefault
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			module = t
		case bool:
			color = t
		case io.Writer:
			out = t
		case Lvl:
			level = t
		default:
			return *defaultLogger, fmt.Errorf("%s:\t%s", "invalid argument", t)
		}
	}
	//newWorker.setLogLevel(level)
	return Logger{
		Module: module,
		worker: newWorker("", format, timeformat, 0, color, out, level),
	}, nil
}

// Logger class that is an interface to user to log messages, Module is the module for which we are testing
// worker is variable of Worker class that is used in bottom layers to log the message
type Logger struct {
	Module string
	worker *worker
}

// Output ...
func (l *Logger) Output(calldepth int, s string) error {
	return l.worker.Minion.Output(calldepth, s)
}

// SetOutput ...
func (l *Logger) SetOutput(w io.Writer) {
	l.worker.Minion.SetOutput(w)
}

//Log  The log commnand is the function available to user to log message, lvl specifies
// the degree of the message the user wants to log, message is the info user wants to log
func (l *Logger) Log(lvl Lvl, message interface{}) {
	l.logInternal(lvl, message, 2)
}

//Log  The log commnand is the function available to user to log message, lvl specifies
// the degree of the message the user wants to log, message is the info user wants to log
func Log(lvl Lvl, message string) {
	defaultLogger.logInternal(lvl, message, 2)
}

func (l *Logger) logInternal(lvl Lvl, message interface{}, pos int) {
	//var formatString string = "#%d %s [%s] %s:%d ▶ %.3s %s"
	_, filename, line, _ := runtime.Caller(pos)
	filename = path.Base(filename)
	info := &info{
		ID:       atomic.AddUint32(&logNo, 1),
		Time:     time.Now().Format(l.worker.timeFormat),
		Module:   l.Module,
		Filename: filename,
		Line:     line,
		Level:    lvl,
		Message:  message,
		//format:   formatString,
	}
	l.worker.log(lvl, 2, info)
}

// Log is Function of Worker class to log a string based on level
func (w *worker) log(level Lvl, calldepth int, info *info) error {
	if w.level < level {
		return nil
	}
	if w.Color {
		buf := &bytes.Buffer{}
		buf.Write(Levels[level].escapedBytes)
		buf.Write([]byte(info.output(w.format, w.spFormat)))
		buf.Write(ansi.Controls["Reset"].Bytes)
		return w.Minion.Output(calldepth+1, buf.String())
	}
	return w.Minion.Output(calldepth+1, info.output(w.format, w.spFormat))
}

// Output Returns formatted string
func (r *info) output(format, spFormat string) string {
	var out string
	switch spFormat {
	/*
		case "yaml":
			l := &info{
				ID:       r.ID,
				Time:     time.Now().String(),
				Module:   r.Module,
				Level:    r.Level,
				Line:     r.Line,
				Filename: r.Filename,
				Message:  r.Message,
			}
			bout, _ := yaml.Marshal(l)
			out = string(bout)
	*/
	case "json":
		var imported interface{}
		switch t := r.Message.(type) {
		case string:
			if err := json.Unmarshal([]byte(t), &imported); err == nil {
				r.Message = imported
			}
		case []byte:
			if err := json.Unmarshal(t, &imported); err == nil {
				r.Message = imported
			}
		}
		bout, _ := json.Marshal(r)
		out = string(bout)
	default:
		out = fmt.Sprintf(
			format,
			r.ID,                  // %[1] // %{id}
			r.Time,                // %[2] // %{time[:fmt]}
			r.Module,              // %[3] // %{module}
			r.Filename,            // %[4] // %{filename}
			r.Line,                // %[5] // %{line}
			Levels[r.Level].Str,   // %[6] // %{level}
			r.Message,             // %[7] // %{message}
			Levels[r.Level].emoji, // %[8] // %{emoji}
		)
		// Ignore printf errors if len(args) > len(verbs)
		if i := strings.LastIndex(out, "%!(EXTRA"); i != -1 {
			return out[:i]
		}
	}
	return out
}
