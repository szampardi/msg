package msg

import (
	"fmt"

	"github.com/nexus166/msg/ansi"
	"github.com/nexus166/msg/unicode"
)

// Lvl is a verbosity level
type Lvl int

// Log Level
const (
	LCrit Lvl = iota
	LErr
	LWarn
	LNotice
	LInfo
	LDebug
	LDefault = LNotice
)

// Level struct
type Level struct {
	ID           int    // export this for comamndlines etc
	Str          string // and this
	emoji        string
	escaped      string
	escapedBytes []byte
}

func initLvl(id int, name, color string, emoji int) Level {
	return Level{
		ID:           id,
		Str:          name,
		emoji:        unicode.CodepageIntToEmoji(emoji),
		escaped:      ansi.Colors[color].Str,
		escapedBytes: ansi.Colors[color].Bytes,
	}
}

var (
	// Levels map all levels to their stuff (a color). also i spent a lot of time deciding these defaults
	Levels = map[Lvl]Level{
		LCrit: initLvl(
			1,
			"FATAL",
			"Red",
			128557, // 😭
		),
		LErr: initLvl(
			2,
			"ERROR",
			"Magenta",
			128545, // 😡
		),
		LWarn: initLvl(
			3,
			"WARN",
			"Yellow",
			128548, // 😤
		),
		LNotice: initLvl(
			4,
			"NOTICE",
			"Green",
			128516, // 😄
		),
		LInfo: initLvl(
			5,
			"INFO",
			"Cyan",
			128523, // 😋
		),
		LDebug: initLvl(
			6,
			"DEBUG",
			"White",
			128533, // 😕
		),
	}
)

// GetLevel retrieve info about a level
func (L Level) GetLevel() (int, string) {
	return L.ID, L.Str
}

// AddLevel to create a new level
func AddLevel(id int, name, color string, emoji int) error {
	if id > 7 {
		Levels[Lvl(id)] = initLvl(id, name, color, emoji)
		return nil
	}
	return fmt.Errorf("%s", "overriding existing log levels is not permitted, pick an id > 7")
}
