package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

func TestLogoColorsSupported_NeedsAtLeast256Colors(t *testing.T) {
	orig := lipgloss.ColorProfile()
	t.Cleanup(func() { lipgloss.SetColorProfile(orig) })

	for profile, want := range map[termenv.Profile]bool{
		termenv.TrueColor: true,
		termenv.ANSI256:   true,
		termenv.ANSI:      false,
	} {
		lipgloss.SetColorProfile(profile)
		if got := logoColorsSupported(); got != want {
			t.Errorf("profile %v: logoColorsSupported() = %v, want %v", profile, got, want)
		}
	}
}

func TestPlayLogo_HidesCursorWhileAnimating(t *testing.T) {
	var out bytes.Buffer
	playLogo(&out, []string{"a"}, func(time.Duration) {})

	got := out.String()
	if !strings.HasPrefix(got, "\x1b[?25l") {
		t.Errorf("animation must hide the cursor before the first frame; starts with %q", got[:min(len(got), 12)])
	}
	if !strings.HasSuffix(got, "\x1b[?25h") {
		t.Errorf("animation must show the cursor again at the end; ends with %q", got[max(0, len(got)-12):])
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
