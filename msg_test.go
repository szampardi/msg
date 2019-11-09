/*

Copyright (c) 2019, SILVANO ZAMPARDI
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

*/

package msg

import (
	"bytes"
	"os"
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
	var worker *worker = newWorker("testing", activeFormat, activeTimeFormat, 0, true, os.Stderr, LDebug)
	if worker.Minion == nil {
		t.Errorf("Minion was not established")
	}
}

func BenchmarkNewWorker(b *testing.B) {
	for n := 0; n <= b.N; n++ {
		worker := newWorker("testing", activeFormat, activeTimeFormat, 0, true, os.Stderr, LDebug)
		if worker == nil {
			panic("Failed to initiate worker")
		}
	}
}
func TestSetFormat(t *testing.T) {
	activeFormat = "%{module} %{lvl} %{message}"
	var buf bytes.Buffer
	log, err := New(StdFormat, DefTimeFmt, "pkgname", &buf)
	log.SetFormat(activeFormat)
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
	//}
	for fmtname, frmt := range Formats {
		for _, test := range tests {
			log, _ := New(frmt.String(), activeTimeFormat, "testing", os.Stdout, LCrit)
			log.SetLogLevel(test.level)
			_fmtname := string(fmtname)
			log.Critical("Log Critical with format " + _fmtname)
			log.Error("Log Error with format " + _fmtname)
			log.Warning("Log Warning with format " + _fmtname)
			log.Notice("Log Notice with format " + _fmtname)
			log.Info("Log Info with format " + _fmtname)
			log.Debug("Log Debug with format " + _fmtname)
			// Count output lines from logger
			all := buf.String()
			buf.Reset()
			t.Log(all)
		}
	}
}
