// COPYRIGHT (c) 2019-2021 SILVANO ZAMPARDI, ALL RIGHTS RESERVED.
// The license for these sources can be found in the LICENSE file in the root directory of this source tree.

package log

import (
	"strings"
	"time"
)

// Format is a format :D
type Format struct {
	Name    string
	_String string
}

func (f Format) String() string {
	return Formats[f.Name]._String
}

const ( //  https://golang.org/src/time/format.go
	// CLITimeFmt command line interface
	CLITimeFmt string = "rfc822"
	// DefTimeFmt default time format
	DefTimeFmt string = "rfc3339"
	// DetailedTimeFmt has nanoseconds
	DetailedTimeFmt string = "rfc3339Nano"
	// CLIFormat  command line interface
	CLIFormat string = "cli"
	// PlainFormat just print message
	PlainFormat string = "plain"
	// PlainFormatWithEmoji just print message with emoji
	PlainFormatWithEmoji string = "plain-emoji"
	// StdFormat standard format
	StdFormat string = "std"
	// StdFormatWithEmoji same as above with an emoji
	StdFormatWithEmoji string = "std-emoji"
	// SimpleFormat short
	SimpleFormat string = "simple"
	// JSONFormat initial attempt at supporting json
	JSONFormat string = "json"
)

var (
	// Formats exported formats
	Formats = map[string]Format{
		CLITimeFmt: {
			Name:    CLITimeFmt,
			_String: time.RFC822,
		},
		DefTimeFmt: {
			Name:    DefTimeFmt,
			_String: time.RFC3339,
		},
		DetailedTimeFmt: {
			Name:    DetailedTimeFmt,
			_String: time.RFC3339Nano,
		},
		CLIFormat: {
			Name:    CLIFormat,
			_String: "%[3]s\t%[2]s\n\t%[7]s\n\n",
		},
		PlainFormat: {
			Name:    PlainFormat,
			_String: "%[7]s",
		},
		PlainFormatWithEmoji: {
			Name:    PlainFormatWithEmoji,
			_String: "%[8]s\t%[7]s",
		},
		StdFormat: {
			Name:    StdFormat,
			_String: "#%[1]d|%[2]s|%[4]s:%[5]d:%[3]s\t%.5[6]s\t%[7]s",
		},
		StdFormatWithEmoji: {
			Name:    StdFormatWithEmoji,
			_String: "#%[1]d|%[2]s|%[4]s:%[5]d:%[3]s\t%[8]s\t%.5[6]s\t%[7]s",
		},
		SimpleFormat: {
			Name:    SimpleFormat,
			_String: "#%[2]s\t%[3]s\t%[7]s",
		},
		JSONFormat: {
			Name:    JSONFormat,
			_String: `{"id":"%[1]d","time":"%[2]s","module":"%[3]s", "line":"%[4]s:%[5]d","message":"%[7]s","level":"%[6]s"}`,
		},
	}
)

var (
	logNo            uint32
	activeFormat     string = Formats[PlainFormat]._String
	activeTimeFormat string = Formats[DefTimeFmt]._String
)

// SetDefaultFormat used to
func SetDefaultFormat() {
	activeFormat, activeTimeFormat = parseFormat(Formats[CLIFormat]._String)
}

func (w *worker) setFormat(format, timeformat string) {
	w.format, w.timeFormat = format, timeformat
}

// SetFormat ...
func (l *Logger) SetFormat(format string) {
	activeFormat, activeTimeFormat = parseFormat(format)
	l.worker.setFormat(activeFormat, activeTimeFormat)
}

func (w *worker) setLogLevel(level Lvl) {
	w.level = level
}

// SetLogLevel to change verbosity
func (l *Logger) SetLogLevel(level Lvl) {
	l.worker.level = level
}

var (
	fmtPlaceholders = map[string]string{
		"%{id}":       "%[1]d",
		"%{time}":     "%[2]s",
		"%{module}":   "%[3]s",
		"%{filename}": "%[4]s",
		"%{file}":     "%[4]s",
		"%{line}":     "%[5]d",
		"%{level}":    "%[6]s",
		"%{lvl}":      "%.3[6]s",
		"%{message}":  "%[7]s",
		//"%{emoji}":  "%[8]s", // added after
	}
)

// Analyze and represent format string as printf format string and time format
func parseFormat(format string) (msgfmt, timefmt string) {
	if len(format) < 10 /* (len of "%{message} */ {
		return activeFormat, activeTimeFormat
	}
	timefmt = activeTimeFormat
	idx := strings.IndexRune(format, '%')
	for idx != -1 {
		msgfmt += format[:idx]
		format = format[idx:]
		if len(format) > 2 {
			if format[1] == '{' {
				// end of curr verb pos
				if jdx := strings.IndexRune(format, '}'); jdx != -1 {
					// next verb pos
					idx = strings.Index(format[1:], "%{")
					// incorrect verb found ("...%{wefwef ...") but after
					// this, new verb (maybe) exists ("...%{inv %{verb}...")
					if idx != -1 && idx < jdx {
						msgfmt += "%%"
						format = format[1:]
						continue
					}
					// get verb and arg
					verb, arg := ph2verb(format[:jdx+1])
					msgfmt += verb
					// check if verb is time
					// here you can handle args for other verbs
					if verb == `%[2]s` && arg != "" /* %{time} */ {
						timefmt = arg
					}
					format = format[jdx+1:]
				} else {
					format = format[1:]
				}
			} else {
				msgfmt += "%%"
				format = format[1:]
			}
		}
		idx = strings.IndexRune(format, '%')
	}
	msgfmt += format
	return
}

// translate format placeholder to printf verb and some argument of placeholder
// (now used only as time format)
func ph2verb(ph string) (verb string, arg string) {
	n := len(ph)
	if n < 4 {
		return ``, ``
	}
	if ph[0] != '%' || ph[1] != '{' || ph[n-1] != '}' {
		return ``, ``
	}
	idx := strings.IndexRune(ph, ':')
	if idx == -1 {
		return fmtPlaceholders[ph], ``
	}
	verb = fmtPlaceholders[ph[:idx]+"}"]
	arg = ph[idx+1 : n-1]
	return
}

// Info class, Contains all the info on what has to logged, time is the current time, Module is the specific module
// For which we are logging, level is the state, importance and type of message logged,
// Message contains the string to be logged, format is the format of string to be passed to sprintf
type info struct {
	ID       uint32      `json:"id"`
	Time     string      `json:"time"`
	Module   string      `json:"module"`
	Level    Lvl         `json:"level"`
	Line     int         `json:"line,omitempty"`
	Filename string      `json:"filename,omitempty"`
	Message  interface{} `json:"message"`
	Emoji    string      `json:"-"`
	//format   string
}
