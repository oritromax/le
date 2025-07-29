package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sjsakib/gudam/utils"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/terminal"
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

	s.PrintUrl()

	addr := fmt.Sprintf(":%d", s.Port)
	err := http.ListenAndServe(addr, nil)

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
	log.Printf("File server is running on  %s", url)

	qrc, err := qrcode.New(url)

	if err != nil {
		log.Printf("error generating QR code: %s", err)
		return
	}

	w := terminal.New()
	if err := qrc.Save(w); err != nil {
		log.Printf("error writing QR code to terminal: %s", err)
		return
	}
}
