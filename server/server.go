package server

import (
	"fmt"
	"os"

	"log"
	"net/http"

	"github.com/sjsakib/gudam/pkg/utils"

	"github.com/mdp/qrterminal/v3"
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
	handler := newHandler(s.Dir)

	s.PrintUrl()

	addr := fmt.Sprintf(":%d", s.Port)
	err := http.ListenAndServe(addr, handler)

	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func (s *Server) PrintUrl() {
	localIP, err := utils.GetLocalIP()
	if err != nil {
		log.Printf("error getting local IP: %s", err)
		localIP = "localhost"
	}

	url := fmt.Sprintf("http://%s:%d", localIP, s.Port)
	log.Printf("Serving files from: %s", s.Dir)
	log.Printf("File server is running on  %s", url)

	qrConfig := qrterminal.Config{
		HalfBlocks: true,
		Level:      qrterminal.M,
		Writer:     os.Stdout,
	}

	qrterminal.GenerateWithConfig(url, qrConfig)

	fmt.Println("")
}
