package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	server *http.Server
	log    *log.Logger
}

func main() {
	err := runServer()
	if err == nil {
		log.Println("finished clean")
		os.Exit(0)
	} else {
		log.Printf("Got error: %v", err)
		os.Exit(1)
	}
}

func runServer() error {
	httpServer := newHTTPServer()

	quit := make(chan os.Signal, 1)

	// listen for interrupt signals, send to quit channel
	signal.Notify(quit,
		os.Interrupt,    // interrupt = SIGINT = Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGUSR1,
	)

	// listen to quit channel, tell server to shutdown
	go func() {
		//cleanup: on interrupt shutdown webserver
		<-quit
		err := httpServer.server.Shutdown(context.Background())
		fmt.Println("Woohoo! Look at")
		if err != nil {
			httpServer.log.Printf("An error occurred on shutdown: %v", err)
		}
	}()

	// listen and serve until error or shutdown is called
	if err := httpServer.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func newHTTPServer() *Server {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":8080"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	http.HandleFunc("/", HealthCheck)
	logger := log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Printf("HTTP server serving at %s", ":8080")
	return &Server{httpServer, logger}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(200)
}
