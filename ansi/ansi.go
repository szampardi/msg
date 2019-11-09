/*

Copyright (c) 2019, SILVANO ZAMPARDI
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

*/

package ansi

import (
	"fmt"
	"strings"
)

const escapePrefix = "\033["

// initialize exported consts and vars that will be used

type ansiSet struct {
	Effect string
	Str    string
	Bytes  []byte
}

func initControl(s, b string) ansiSet {
	return ansiSet{
		Effect: s,
		Bytes:  []byte(b),
		Str:    b,
	}
}

// https://rosettacode.org/wiki/Terminal_control/Coloured_text#ANSI_escape_codes
const (
	Black int = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// ANSI holds info about an escape. Content is stored in memory accessible by function
type colorSet struct {
	Code    int
	Str     string
	Bytes   []byte
	Bgcode  int
	Bgstr   string
	BgBytes []byte
}

func initColor(c int) colorSet {
	intStr := func(c int) string {
		return fmt.Sprintf("%d", c)
	}
	esc, bgesc := escapePrefix+intStr(c)+"m", escapePrefix+intStr(c+10)+"m"
	escB, bgescB := []byte(esc), []byte(bgesc)
	return colorSet{
		Code:    c,
		Str:     esc,
		Bytes:   escB,
		Bgcode:  c + 10, // background
		Bgstr:   bgesc,
		BgBytes: bgescB,
	}
}

var (
	// Colors are predeclared and accessible externally
	Colors = map[string]colorSet{
		"Black":   initColor(Black),
		"Red":     initColor(Red),
		"Green":   initColor(Green),
		"Yellow":  initColor(Yellow),
		"Blue":    initColor(Blue),
		"Magenta": initColor(Magenta),
		"Cyan":    initColor(Cyan),
		"White":   initColor(White),
	}
	// Controls exports some other strings/bytes
	Controls = map[string]ansiSet{
		"Prefix": initControl(
			"Prefix",
			escapePrefix,
		),
		"Bold": initControl(
			"Bold",
			escapePrefix+"1m",
		),
		"Clear": initControl(
			"Clear",
			escapePrefix+"2J",
		),
		"Dim": initControl(
			"Dim",
			escapePrefix+"2m",
		),
		"Hide": initControl(
			"Hide",
			escapePrefix+"8m",
		),
		"Blink": initControl(
			"Blink",
			escapePrefix+"5m",
		),
		"Unblink": initControl(
			"Unblink",
			escapePrefix+"25m",
		),
		"Reset": initControl(
			"Reset",
			escapePrefix+"0m",
		),
		"Reverse": initControl(
			"Reverse",
			escapePrefix+"7m",
		),
		"Underline": initControl(
			"Underline",
			escapePrefix+"4m",
		),
	}
)

// GetColor returns an ansi with the inmem material ready to be printed/io.Written.
// Default is Red
func GetColor(color string) colorSet {
	if x := Colors[color]; x.Str != `` {
		return x
	}
	return initColor(Red)
}

// PaintStrings accepts strings and returns a colored string ready to be printed.
// Bg is for background
func PaintStrings(color string, bg bool, sep string, s ...string) string {
	if ansi := GetColor(color); ansi.Str != `` {
		all := strings.Join(s, sep)
		if bg {
			return ansi.Bgstr + all
		}
		return ansi.Str + all
	}
	return ``
}
