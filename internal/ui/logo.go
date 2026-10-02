package ui

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"skillshare/internal/theme"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

const (
	// logoMinWidth is the narrowest terminal that shows the logo; below it
	// only the text lines are printed.
	logoMinWidth = 60
	logoGap      = "    "
	logoIndent   = "  "
	logoSteps    = 16
	logoFrameGap = 55 * time.Millisecond
)

type rgb struct{ r, g, b float64 }

// LogoBanner prints the hand-drawn logo with text lines beside it, centered
// vertically. On a terminal it first plays a short "light up" animation that
// spreads color out from the bulb. Without a terminal, without color, or on a
// narrow terminal it prints only the text lines.
func LogoBanner(lines []string, animate bool) {
	if !logoFits() {
		for _, line := range lines {
			fmt.Println(line)
		}
		return
	}
	if !animate {
		fmt.Print(strings.Join(logoFrame(1, lines, true), "\n") + "\n")
		return
	}
	rows := len(logoPixels) / 2
	for step := 0; step <= logoSteps; step++ {
		if step > 0 {
			fmt.Printf("\x1b[%dA", rows)
		}
		progress := easeOut(float64(step) / logoSteps)
		// Text fades in over the last few frames and settles on the final one.
		showText := step >= logoSteps-4
		frame := logoFrame(progress, lines, showText)
		if showText && step < logoSteps {
			frame = logoFrame(progress, dimLines(lines), true)
		}
		for _, line := range frame {
			fmt.Print(line + "\x1b[K\n")
		}
		if step < logoSteps {
			time.Sleep(logoFrameGap)
		}
	}
}

func logoFits() bool {
	t := theme.Get()
	if t.NoColor || t.Plain {
		return false
	}
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	return err == nil && width >= logoMinWidth
}

// logoFrame renders the logo at the given animation progress (0..1) with
// the text lines placed in a column to its right.
func logoFrame(progress float64, lines []string, showText bool) []string {
	colors := logoColors(progress)
	rows := len(colors) / 2
	top := (rows - len(lines)) / 2
	out := make([]string, rows)
	for row := 0; row < rows; row++ {
		var b strings.Builder
		b.WriteString(logoIndent)
		upper, lower := colors[row*2], colors[row*2+1]
		for x := range upper {
			b.WriteString(halfBlock(upper[x], lower[x]))
		}
		if i := row - top; showText && i >= 0 && i < len(lines) && lines[i] != "" {
			b.WriteString(logoGap + lines[i])
		}
		out[row] = b.String()
	}
	return out
}

// logoColors returns each pixel's color at the given progress, nil for
// transparent pixels. Pixels start as a faint grey outline and take on
// their real color as a ring spreads out from the center.
func logoColors(progress float64) [][]*rgb {
	bg := rgb{18, 19, 22}
	if theme.Get().Mode == theme.ModeLight {
		bg = rgb{250, 250, 250}
	}
	n := len(logoPixels)
	center := float64(n-1) / 2
	maxDist := math.Hypot(center, center)
	out := make([][]*rgb, n)
	for y, row := range logoPixels {
		out[y] = make([]*rgb, len(row)/6)
		for x := range out[y] {
			c, ok := parseHex(row[x*6 : x*6+6])
			if !ok {
				continue
			}
			lum := 0.3*c.r + 0.59*c.g + 0.11*c.b
			grey := mix(bg, rgb{lum, lum, lum}, 0.4)
			dist := math.Hypot(float64(x)-center, float64(y)-center) / maxDist
			k := math.Max(0, math.Min(1, (progress*1.25-dist)/0.2))
			mixed := mix(grey, c, k)
			out[y][x] = &mixed
		}
	}
	return out
}

func halfBlock(upper, lower *rgb) string {
	switch {
	case upper != nil && lower != nil:
		return lipgloss.NewStyle().Foreground(upper.color()).Background(lower.color()).Render("▀")
	case upper != nil:
		return lipgloss.NewStyle().Foreground(upper.color()).Render("▀")
	case lower != nil:
		return lipgloss.NewStyle().Foreground(lower.color()).Render("▄")
	default:
		return " "
	}
}

func dimLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = theme.Dim().Render(StripANSI(line))
	}
	return out
}

func parseHex(s string) (rgb, bool) {
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return rgb{}, false
	}
	return rgb{float64(v >> 16 & 0xff), float64(v >> 8 & 0xff), float64(v & 0xff)}, true
}

func mix(a, b rgb, k float64) rgb {
	return rgb{a.r + (b.r-a.r)*k, a.g + (b.g-a.g)*k, a.b + (b.b-a.b)*k}
}

func (c rgb) color() lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", int(c.r), int(c.g), int(c.b)))
}

func easeOut(t float64) float64 { return 1 - math.Pow(1-t, 3) }
