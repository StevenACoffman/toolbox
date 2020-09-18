package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestWaiter(t *testing.T) {
	t.Run("Wait with func", func(t *testing.T) {
		var finished bool
		// Get the operating system process
		proc, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		// Discard noisy logs
		logger := log.New(ioutil.Discard, "", log.LstdFlags)
		go func() {
			Waiter(logger)
			finished = true
		}()
		// if we signal too early, Waiter isn't listening yet
		time.Sleep(10 * time.Millisecond)
		// Send the SIGQUIT
		proc.Signal(syscall.SIGQUIT)
		// if we test finished too early, finished may not have been updated yet
		time.Sleep(10 * time.Millisecond)
		// reset signal notification
		signal.Reset()
		if !finished {
			t.Error("Waiter Did Not Exit")
		}
	})
}
