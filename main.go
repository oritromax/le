package main

import (
	"flag"
	"log"

	"go.sakib.dev/le/server"
	"go.sakib.dev/le/tui"
)

func main() {
	dir := flag.String("dir", ".", "Directory to serve files from")
	port := flag.Int("port", 8080, "Port to run the file server on")

	flag.Parse()

	eventCh := make(chan server.ServerEventName, 10)
	srvr, err := server.NewServer(*dir, *port, eventCh)

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	go func() {
		if err := srvr.Start(); err != nil {
			log.Fatalf("Failed to start srvr: %v", err)
		}
	}()

	err = tui.Start(srvr, eventCh)
	if err != nil {
		log.Fatalf("Failed to start TUI: %v", err)
	}
}
