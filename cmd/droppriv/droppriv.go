package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if os.Getuid() == 0 {
		fmt.Println("Dropping privileges...")
		if err := drop(); err != nil {
			fmt.Println("Failed to drop privileges:", err)
			os.Exit(1)
		}
	}

	l, err := net.FileListener(os.NewFile(3, "[socket]"))
	if err != nil {
		// Yell into the void.
		fmt.Println("Failed to listen on FD 3:", err)
		os.Exit(1)
	}

	http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I am process %d running as %d/%d", os.Getpid(), os.Getuid(), os.Getgid())
	}))
}

func drop() error {
	l, err := net.Listen("tcp", ":80")
	if err != nil {
		return err
	}

	f, err := l.(*net.TCPListener).File()
	if err != nil {
		return err
	}

	cmd := exec.Command(os.Args[0])
	cmd.ExtraFiles = []*os.File{f}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: 65534,
			Gid: 65534,
		},
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	fmt.Printf("Spawned process %d, exiting\n", cmd.Process.Pid)
	cmd.Process.Release()
	os.Exit(0)
	return nil /* unreachable */
}
