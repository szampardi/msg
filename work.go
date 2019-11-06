package msg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nexus166/msg/ansi"
)

// Worker class, Worker is a log object used to log messages and Color specifies
// if colored output is to be produced
type worker struct {
	Minion     *log.Logger
	Color      bool
	format     Format
	timeFormat Format
	level      Lvl
}

//NewWorker  Returns an instance of worker class, prefix is the string attached to every log,
// flag determine the log params, color parameters verifies whether we need colored outputs or not
func newWorker(prefix string, format, timeformat Format, flag int, color bool, out io.Writer) *worker {
	return &worker{
		Minion:     log.New(out, prefix, flag),
		Color:      color,
		format:     format,
		timeFormat: timeformat,
	}
}

// New Returns a new instance of logger class, module is the specific module for which we are logging
// , color defines whether the output is to be colored or not, out is instance of type io.Writer defaults
// to os.Stderr
func New(format, timeformat Format, args ...interface{}) (*Logger, error) {
	var module string = Levels[LDefault].Str
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
			return nil, fmt.Errorf("%s:\t%s", "invalid argument", t)
		}
	}
	newWorker := newWorker("", format, timeformat, 0, color, out)
	newWorker.setLogLevel(level)
	return &Logger{Module: module, worker: newWorker}, nil
}

// Logger class that is an interface to user to log messages, Module is the module for which we are testing
// worker is variable of Worker class that is used in bottom layers to log the message
type Logger struct {
	Module string
	worker *worker
}

//Log  The log commnand is the function available to user to log message, lvl specifies
// the degree of the message the user wants to log, message is the info user wants to log
func (l *Logger) Log(lvl Lvl, message string) {
	l.logInternal(lvl, message, 2)
}

func (l *Logger) logInternal(lvl Lvl, message string, pos int) {
	//var formatString string = "#%d %s [%s] %s:%d ▶ %.3s %s"
	_, filename, line, _ := runtime.Caller(pos)
	filename = path.Base(filename)
	info := &info{
		ID:       atomic.AddUint64(&logNo, 1),
		Time:     time.Now().Format(Fmt[l.worker.timeFormat]),
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
		buf.Write([]byte(info.output(Format(w.format))))
		buf.Write(ansi.Controls["Reset"].Bytes)
		return w.Minion.Output(calldepth+1, buf.String())
	}
	return w.Minion.Output(calldepth+1, info.output(Format(w.format)))
}

// Output Returns a proper string to be outputted for a particular info
func (r *info) output(format Format) string {
	msg := fmt.Sprintf(
		Fmt[format],
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
	if i := strings.LastIndex(msg, "%!(EXTRA"); i != -1 {
		return msg[:i]
	}
	return msg
}
