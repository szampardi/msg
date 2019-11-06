package ansi

import (
	"testing"
)

func TestColorInit(t *testing.T) {
	for color := range Colors {
		t.Logf("%v", initColor(Colors[color].Code))
	}
}

func TestValsLoad(t *testing.T) {
	for color := range Colors {
		t.Logf("%s:\t%v", "Found color", color)
	}
}

func TestPrintVals(t *testing.T) {
	for color := range Colors {
		t.Logf("%s%s", Colors[color].Str, color)
		t.Logf("%s", Controls["Reset"].Str)
	}
}

func TestBlinkVals(t *testing.T) {
	for color := range Colors {
		t.Logf("%s", Colors[color].Str+Controls["Blink"].Str+color) // on verbose, weird output (missing parts) should actually be fine
		t.Logf("%s", Controls["Reset"].Str)
	}
}

func TestGetColorSet(t *testing.T) {
	for color := range Colors {
		t.Logf("%v", GetColor(color))
		t.Logf("%s", Controls["Reset"].Str)
	}
	t.Logf("%v", GetColor("Antimatter"))
}

func TestPaintText(t *testing.T) {
	for color := range Colors {
		t.Logf("%s", PaintStrings(color, false, " ", "this is color ", color))
		t.Logf("%s", PaintStrings(color, true, " ", "this is background ", color))
		t.Logf("%s", Controls["Reset"].Str)
	}
}
