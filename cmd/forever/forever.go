package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	logger.Printf("Starting up!")

	// make a quit channel for operating system signals, buffered to size 1
	quit := make(chan os.Signal, 1)

	logger.Printf("Just going to wait here until you press control-C")
	// block, waiting for receive on quit channel
	<-quit
	logger.Printf("I never get here!")
	// 0 means no errors
	os.Exit(0)
}
