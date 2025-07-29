package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sjsakib/gudam/utils"
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

	localIP, err := utils.GetLocalIP()
	if err != nil {
		log.Printf("error getting local IP: %s", err)
		localIP = "localhost"
	}

	log.Printf("File server is running on http://%s:%d", localIP, s.Port)
	addr := fmt.Sprintf(":%d", s.Port)
	err = http.ListenAndServe(addr, nil)

	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}
