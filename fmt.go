package msg

import (
	"strings"
	"time"
)

// Format is a format :D
type Format string

const ( //  https://golang.org/src/time/format.go
	// CLITimeFmt command line interface
	CLITimeFmt Format = "rfc822"
	// DefTimeFmt default time format
	DefTimeFmt Format = "rfc3339"
	// DetailedTimeFmt has nanoseconds
	DetailedTimeFmt Format = "rfc3339Nano"
	// CLIFormat  command line interface
	CLIFormat Format = "cli"
	// StdFormat standard format
	StdFormat Format = "std"
	// StdFormatWithEmoji same as above with an emoji
	StdFormatWithEmoji Format = "std-emoji"
	// SimpleFormat short
	SimpleFormat Format = "simple"
	// JSONFormat initial attempt at supporting json
	JSONFormat Format = "json"
)

var (
	// Fmt exported formats
	Fmt = map[Format]string{
		CLITimeFmt:         time.RFC822,
		DefTimeFmt:         time.RFC3339,
		DetailedTimeFmt:    time.RFC3339Nano,
		CLIFormat:          "%[3]s\t%[2]s\n\t%[7]s\n\n",
		StdFormat:          "#%[1]d|%[2]s|%[4]s:%[5]d\t%.5[6]s\t%[7]s",
		StdFormatWithEmoji: "#%[1]d|%[2]s|%[4]s:%[5]d\t%[8]s\t%.5[6]s\t%[7]s",
		SimpleFormat:       "#%[2]s\t%[3]s\t%[7]s",
		JSONFormat:         `{"id":"%[1]d","time":"%[2]s","module":"%[3]s", "line":"%[4]s:%[5]d","message":"%[7]s","level":"%[6]s"}`, // message still has to be escaped
	}
)

var (
	logNo            uint64
	activeFormat     Format = Format(Fmt[StdFormat])
	activeTimeFormat Format = Format(Fmt[DefTimeFmt])
)

// SetDefaultFormat used to
func SetDefaultFormat() {
	activeFormat, activeTimeFormat = parseFormat(Fmt[CLIFormat])
}

func (w *worker) setFormat(format, timeformat Format) {
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
func parseFormat(format string) (msgfmt, timefmt Format) {
	if len(format) < 10 /* (len of "%{message} */ {
		return activeFormat, activeTimeFormat
	}
	timefmt = activeTimeFormat
	idx := strings.IndexRune(format, '%')
	for idx != -1 {
		msgfmt += Format(format[:idx])
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
					msgfmt += Format(verb)
					// check if verb is time
					// here you can handle args for other verbs
					if verb == `%[2]s` && arg != "" /* %{time} */ {
						timefmt = Format(arg)
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
	msgfmt += Format(format)
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
	ID       uint64
	Time     string
	Module   string
	Level    Lvl
	Line     int
	Filename string
	Message  string
	Emoji    string
	//format   string
}
