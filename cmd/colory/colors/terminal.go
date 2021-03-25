package colors

// terminal.go file has terminal specific functions
import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// IsTerminal return true if the file descriptor is terminal.
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), tcgetattr)
	return err == nil
}

// termStatusReport attempts to find various facts about the terminal
// using xterm control sequences like this:
//
// \e]11;?\a
// (\e and \a are the ESC and BEL characters, respectively.)
//
// Xterm-compatible terminals should reply with the same sequence,
// with the question mark replaced by an X11 color name
// e.g. rgb:0000/0000/0000 for black.
//
// This will *probably* not work on Windows. But who knows these days?
func termStatusReport(sequence int) (string, error) {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "screen") {
		return "", fmt.Errorf("Invalid Terminal Status")
	}

	t, err := unix.IoctlGetTermios(unix.Stdout, tcgetattr)
	if err != nil {
		return "", fmt.Errorf("Invalid Terminal Status")
	}
	defer func() {
		err = unix.IoctlSetTermios(unix.Stdout, tcsetattr, t)
	}()

	noecho := *t
	noecho.Lflag &^= unix.ECHO
	noecho.Lflag &^= unix.ICANON
	err = unix.IoctlSetTermios(unix.Stdout, tcsetattr, &noecho)
	if err != nil {
		return "", fmt.Errorf("Invalid Terminal Status")
	}

	fmt.Printf("\033]%d;?\033\\", sequence) // nolint:ka-banned-symbol // we are not in a web app
	s, ok := readWithTimeout(os.Stdout)
	if !ok {
		return "", fmt.Errorf("Invalid Terminal Status")
	}
	return s, err
}

// Log is a simple terminal console log to stdout
// This is not intended for use in a web application
func Log(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

// Err is a simple terminal console log to Stderr
// This is not intended for use in a web application
func Err(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

// StyledLog returns the string rendered with all styles.
func StyledLog(styles []string, format string, args ...interface{}) {
	ss := Style{
		string: fmt.Sprintf(format, args...),
		styles: styles,
	}
	fmt.Fprint(os.Stdout, ss.String())
}

// StyledErr returns the string rendered with all styles.
func StyledErr(styles []string, format string, args ...interface{}) {
	ss := Style{
		string: fmt.Sprintf(format, args...),
		styles: styles,
	}
	fmt.Fprint(os.Stderr, ss.String())
}

func GreenLog(format string, args ...interface{}) {
	StyledLog([]string{GreenFGStyle}, format, args...)
}

func BoldLog(format string, args ...interface{}) {
	StyledLog([]string{BoldSeq}, format, args...)
}

func BoldGreenLog(format string, args ...interface{}) {
	StyledLog([]string{GreenFGStyle, BoldSeq}, format, args...)
}

func RedErrLog(format string, args ...interface{}) {
	StyledErr([]string{YellowFGStyle, BoldSeq}, format, args...)
}
