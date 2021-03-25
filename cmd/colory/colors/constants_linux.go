package colors

// constants_linux.go contains platform specific constants
import "golang.org/x/sys/unix"

// terminal ioctl commands
const (
	tcgetattr = unix.TCGETS // Get the terminal properties in a termios structure
	tcsetattr = unix.TCSETS // Set terminal properties from a termios structure
)