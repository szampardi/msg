package msg

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/nexus166/msg/ansi"
)

func BenchmarkLoggerLog(b *testing.B) {
	b.StopTimer()
	log, err := New(StdFormat, DefTimeFmt, "testing")
	if err != nil {
		panic(err)
	}
	var tests = []struct {
		level   Lvl
		message string
	}{{LCrit, "Critical Logging"},
		{LErr, "Error logging"},
		{LWarn, "Warning logging"},
		{LNotice, "Notice Logging"},
		{LInfo, "Info Logging"},
		{LDebug, "Debug logging"}}
	b.StartTimer()
	for _, test := range tests {
		for n := 0; n <= b.N; n++ {
			log.Log(test.level, test.message)
		}
	}
}

func BenchmarkLoggerNew(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		log, err := New(StdFormat, DefTimeFmt, "testing")
		if err != nil && log == nil {
			panic(err)
		}
	}
}

func TestLoggerNew(t *testing.T) {
	log, err := New(StdFormat, DefTimeFmt, "testing")
	if err != nil {
		panic(err)
	}
	if log.Module != "testing" {
		t.Errorf("Unexpected module: %s", log.Module)
	}
}

func TestColorString(t *testing.T) {
	colorCode := ansi.PaintStrings(`Black`, true, " ")
	if colorCode != "\033[40m" {
		t.Errorf("Unexpected string: %s", colorCode)
	}
}

func TestNewWorker(t *testing.T) {
	var worker *worker = newWorker("testing", activeFormat, activeTimeFormat, 0, true, os.Stderr)
	if worker.Minion == nil {
		t.Errorf("Minion was not established")
	}
}

func BenchmarkNewWorker(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		worker := newWorker("testing", activeFormat, activeTimeFormat, 0, true, os.Stderr)
		if worker == nil {
			panic("Failed to initiate worker")
		}
	}
}
func TestSetFormat(t *testing.T) {
	activeFormat = "%{module} %{lvl} %{message}"
	var buf bytes.Buffer
	log, err := New(StdFormat, DefTimeFmt, "pkgname", &buf)
	log.SetFormat(Fmt[activeFormat])
	if err != nil || log == nil {
		panic(err)
	}
	log.Infof("Test %d", 123)
	want := "pkgname FAT Test 123\n"
	have := buf.String()
	if want != have {
		t.Errorf("\nWant: %sHave: %s", want, have)
	}
}

func TestLogLevel(t *testing.T) {
	var tests = []struct {
		level   Lvl
		message string
	}{{LCrit, "Critical Logging"},
		{LErr, "Error logging"},
		{LWarn, "Warning logging"},
		{LNotice, "Notice Logging"},
		{LInfo, "Info Logging"},
		{LDebug, "Debug logging"}}
	var buf bytes.Buffer
	log, err := New(StdFormat, DefTimeFmt, "testing")
	if err != nil {
		panic(err)
	}
	for i, test := range tests {
		log.SetLogLevel(test.level)
		log.Critical("Log Critical")
		log.Error("Log Error")
		log.Warning("Log Warning")
		log.Notice("Log Notice")
		log.Info("Log Info")
		log.Debug("Log Debug")
		// Count output lines from logger
		all := buf.String()
		buf.Reset()
		count := strings.Count(all, "\n")
		if i+1 != count {
			t.Error()
		}
		t.Log(all)
	}
}
