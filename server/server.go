package server

import (
	"fmt"
	"os"

	"log/slog"
	"net/http"

	"go.sakib.dev/le/logger"
	"go.sakib.dev/le/pkg/utils"

	"github.com/mdp/qrterminal/v3"
)

type Server struct {
	Dir  string
	Port int
}

func NewServer(dir string, port int) *Server {
	slog.SetDefault(slog.New(logger.NewHandler()))

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
		slog.Error("Error getting local IP", "error", err)
		localIP = "localhost"
	}

	url := fmt.Sprintf("http://%s:%d", localIP, s.Port)
	slog.Info("Serving files from", "directory", s.Dir)
	slog.Info("File server is running on", "url", url)

	qrConfig := qrterminal.Config{
		HalfBlocks: true,
		Level:      qrterminal.M,
		Writer:     os.Stdout,
	}

	qrterminal.GenerateWithConfig(url, qrConfig)

	fmt.Println("")
}
