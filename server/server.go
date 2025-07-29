package server

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Dir  string
	Port int
}

func NewServer(dir string, port int) *Server {
	return &Server{
		Dir:  dir,
		Port: port,
	}
}

func (s *Server) Start() error {
	fs := http.FileServer(http.Dir(s.Dir))
	http.Handle("/", fs)

	addr := fmt.Sprintf(":%d", s.Port)
	log.Printf("File server is running on http://localhost:%d", s.Port)
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}
