package main

import (
	"flag"
	"log"

	"go.sakib.dev/le/server"
)


func main() {
	dir := flag.String("dir", ".", "Directory to serve files from")
	port := flag.Int("port", 8080, "Port to run the file server on")

	flag.Parse()

	server := server.NewServer(*dir, *port)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
