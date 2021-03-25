// colors package to keep track of
// ANSI color codes for command line tools.
// Windows does not support them,
// so we empty them out on that platform.
//
// Some people have terminals with dark backgrounds, while
// Some people have terminals with light backgrounds.
// We try to detect that and set high contrast default colors.
//
// End users may enjoy BoldGreenLog or RedErrLog
package colors

// colors.go file has terminal color manipulation machinery
import (
	"bytes"
	"fmt"
	"image/color"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	// simple foreground ANSI Color sequences
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	Gray      = "\033[37m"
	White     = "\033[97m"
	Bold      = "\033[1m"
	Underline = "\033[4m"
	// useful package variables
	IsUsingDarkBackground = false
	RedFGStyle            = ""
	GreenFGStyle          = ""
	YellowFGStyle         = ""
	BlueFGStyle           = ""
	MagentaFGStyle        = ""
	CyanFGStyle           = ""
	GrayFGStyle           = ""
	WhiteFGStyle          = ""
)

const (
	CSI        = "\x1b[" // ANSI Control Sequence Introducer (CSI) sequences
	Foreground = "38"    // ANSI Code for Set Foreground color
	Background = "48"    // ANSI Code for Set background color
)

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Magenta = ""
		Cyan = ""
		Gray = ""
		White = ""
		return
	}
	IsUsingDarkBackground = HasDarkBackground()
	if IsUsingDarkBackground {
		RedFGStyle = ANSIBrightRed.Sequence(false)
		GreenFGStyle = RGBColor("#00ff00").Sequence(false)
		YellowFGStyle = RGBColor("#ffff00").Sequence(false)
		BlueFGStyle = RGBColor("#000080").Sequence(false)    // low contrast
		MagentaFGStyle = RGBColor("#ff00ff").Sequence(false) // low contrast
		CyanFGStyle = RGBColor("#00ffff").Sequence(false)
		GrayFGStyle = RGBColor("#c0c0c0").Sequence(false)
	} else {
		RedFGStyle = RGBColor("#800000").Sequence(false)
		GreenFGStyle = RGBColor("#008000").Sequence(false)
		YellowFGStyle = RGBColor("#808000").Sequence(false)
		BlueFGStyle = RGBColor("#0000ff").Sequence(false) // low contrast
		MagentaFGStyle = RGBColor("#800080").Sequence(false)
		CyanFGStyle = RGBColor("#008080").Sequence(false)
		GrayFGStyle = RGBColor("#808080").Sequence(false)
	}
	Red = Render(RedFGStyle)
	Green = Render(GreenFGStyle)
	Yellow = Render(YellowFGStyle)
	Blue = Render(BlueFGStyle)       // warning: low contrast
	Magenta = Render(MagentaFGStyle) // warning: low contrast
	Cyan = Render(CyanFGStyle)
	Gray = Render(GrayFGStyle)
}

// Render does not Reset the style at the end.
func Render(seq string) string {
	return fmt.Sprintf("%s%sm", CSI, seq)
}

// ForegroundColor returns the terminal's default foreground color.
func ForegroundColor() Color {
	if !IsTerminal(os.Stdout.Fd()) {
		return NoColor{}
	}

	return foregroundColor()
}

// BackgroundColor returns the terminal's default background color.
func BackgroundColor() Color {
	if !IsTerminal(os.Stdout.Fd()) {
		return NoColor{}
	}

	return backgroundColor()
}

// HasDarkBackground returns whether terminal uses a dark-ish background.
func HasDarkBackground() bool {
	b := Brightness(ConvertToRGB(BackgroundColor()))
	return b < 128.0
}

// Brightness [0..255] of a given color - ignoring alpha
// We will use the formula taken from https://www.w3.org/TR/AERT#color-contrast
// Color brightness is determined by the following formula:
// ((Red value X 299) + (Green value X 587) + (Blue value X 114)) / 1000
func Brightness(rgb RGBColor) float64 {
	aColor := parseHexColor(string(rgb))
	// because of
	// http://stackoverflow.com/questions/35374300/why-does-golang-rgba-rgba-method-use-and
	// we cannot use .RGBA() values. It will mess up the final luminance result

	// instead, we will use the colors value from 0..255 for calculation
	red := float64(aColor.R)
	green := float64(aColor.G)
	blue := float64(aColor.B)

	// need to convert uint32 to float64
	return 0.299*red + 0.587*green + 0.114*blue
}

type Color interface {
	Sequence(bg bool) string
}

// Hex returns the hex "html" representation of the color, as in #ff0080.
func (c RGBColor) Hex() string {
	f := parseHexColor(string(c))
	return fmt.Sprintf("#%02x%02x%02x", f.R, f.G, f.B)
}

type NoColor struct{}

func (c NoColor) Sequence(_ bool) string {
	return ""
}

// ANSIColor is a color (0-15) as defined by the ANSI Standard.
type ANSIColor int

// RGBColor is a hex-encoded color, e.g. "#abcdef".
type RGBColor string

func (c RGBColor) Sequence(bg bool) string {
	f := parseHexColor(string(c))

	prefix := Foreground
	if bg {
		prefix = Background
	}
	return fmt.Sprintf("%s;2;%d;%d;%d", prefix, f.R, f.G, f.B)
}

func ConvertToRGB(c Color) RGBColor {
	var hex string
	switch v := c.(type) {
	case RGBColor:
		return c.(RGBColor)
	case ANSIColor:
		hex = ansiHex[v]
	}

	ch := parseHexColor(hex)
	chs := fmt.Sprintf("#%x%x%x", ch.R, ch.G, ch.B)
	return RGBColor(chs)
}

// This method only intended to lookup valid data from our constants
// It should never produce an error, under those circumstances
func parseHexColor(s string) (c color.RGBA) {
	var err error
	c.A = 0xff // default to opaque
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	case 3:
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default: // default to white
		c.R = 0xff
		c.G = 0xff
		c.B = 0xff
	}
	if err != nil { // default to white
		c.R = 0xff
		c.G = 0xff
		c.B = 0xff
	}
	return c
}

func (c ANSIColor) Sequence(bg bool) string {
	col := int(c)
	bgMod := func(c int) int {
		if bg {
			return c + 10
		}
		return c
	}

	if col < 8 {
		return fmt.Sprintf("%d", bgMod(col)+30)
	}
	return fmt.Sprintf("%d", bgMod(col-8)+90)
}

func xTermColor(s string) (RGBColor, error) {
	if len(s) < 24 || len(s) > 25 {
		return "", fmt.Errorf("invalid color")
	}

	switch {
	case strings.HasSuffix(s, "\a"):
		s = strings.TrimSuffix(s, "\a")
	case strings.HasSuffix(s, "\033\\"):
		s = strings.TrimSuffix(s, "\033\\")
	default:
		return "", fmt.Errorf("invalid color")
	}

	s = s[4:]

	prefix := ";rgb:"
	if !strings.HasPrefix(s, prefix) {
		return "", fmt.Errorf("invalid color")
	}
	s = strings.TrimPrefix(s, prefix)

	h := strings.Split(s, "/")
	hex := fmt.Sprintf("#%s%s%s", h[0][:2], h[1][:2], h[2][:2])
	return RGBColor(hex), nil
}

func backgroundColor() Color {
	s, err := termStatusReport(11)
	if err == nil {
		c, err := xTermColor(s)
		if err == nil {
			return c
		}
	}

	// Some cli tools use/set the COLORFGBG env variable
	colorFGBG := os.Getenv("COLORFGBG")
	if strings.Contains(colorFGBG, ";") {
		c := strings.Split(colorFGBG, ";")
		i, err := strconv.Atoi(c[1])
		if err == nil {
			return ANSIColor(i)
		}
	}

	// default black
	return ANSIColor(0)
}

func foregroundColor() Color {
	s, err := termStatusReport(10)
	if err == nil {
		c, err := xTermColor(s)
		if err == nil {
			return c
		}
	}

	colorFGBG := os.Getenv("COLORFGBG")
	if strings.Contains(colorFGBG, ";") {
		c := strings.Split(colorFGBG, ";")
		i, err := strconv.Atoi(c[0])
		if err == nil {
			return ANSIColor(i)
		}
	}

	// default gray
	return ANSIColor(7)
}

func readWithTimeout(f *os.File) (string, bool) {
	var readfds unix.FdSet
	fd := int(f.Fd())
	readfds.Set(fd)

	for {
		// Use select to attempt to read from os.Stdout for 100 ms
		_, err := unix.Select(fd+1, &readfds, nil, nil, &unix.Timeval{Usec: 100000})
		if err == nil {
			break
		}
		// On MacOS we can see EINTR here if the user
		// pressed ^Z. Similar to issue:
		// https://github.com/golang/go/issues/22838
		if runtime.GOOS == "darwin" && err == unix.EINTR {
			continue
		}
		return "", false
	}

	if !readfds.IsSet(fd) {
		// select(read timeout)
		return "", false
	}

	// n > 0 => is readable
	var data []byte
	b := make([]byte, 1)
	for {
		_, err := f.Read(b)
		if err != nil {
			return "", false
		}

		data = append(data, b[0])

		// data sent by terminal is either terminated by BEL (\a) or ST (ESC \)
		if bytes.HasSuffix(data, []byte("\a")) || bytes.HasSuffix(data, []byte("\033\\")) {
			break
		}
	}
	return string(data), true
}
