package ui

import (
	"testing"
)

func TestLogoColors_FinalFrameUsesLogoColors(t *testing.T) {
	x, y := firstDrawnPixel(t)
	want, _ := parseHex(logoPixels[y][x*6 : x*6+6])

	got := logoColors(1)[y][x]
	if got == nil || *got != want {
		t.Errorf("pixel (%d,%d) at progress 1 = %v, want %v", x, y, got, want)
	}
}

func TestLogoColors_FirstFrameIsNotYetColored(t *testing.T) {
	x, y := firstDrawnPixel(t)
	logo, _ := parseHex(logoPixels[y][x*6 : x*6+6])

	got := logoColors(0)[y][x]
	if got == nil || *got == logo {
		t.Errorf("pixel (%d,%d) at progress 0 = %v, want the grey outline, not %v", x, y, got, logo)
	}
}

func TestLogoFrame_HasOneLinePerTwoPixelRows(t *testing.T) {
	if got, want := len(logoFrame(1, []string{"a"}, true)), len(logoPixels)/2; got != want {
		t.Errorf("frame has %d lines, want %d", got, want)
	}
}

func firstDrawnPixel(t *testing.T) (x, y int) {
	t.Helper()
	for y, row := range logoPixels {
		for x := 0; x*6 < len(row); x++ {
			if _, ok := parseHex(row[x*6 : x*6+6]); ok {
				return x, y
			}
		}
	}
	t.Fatal("logo has no drawn pixels")
	return 0, 0
}
