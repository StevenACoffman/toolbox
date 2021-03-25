// +build darwin dragonfly freebsd netbsd openbsd
// +build !solaris
// +build !illumos

package colors

// constants_linux.go contains platform specific constants
import "golang.org/x/sys/unix"

const (
	// terminal ioctl commands
	tcgetattr = unix.TIOCGETA // Get the terminal properties in a termios structure
	tcsetattr = unix.TIOCSETA // Set terminal properties from a termios structure
)


