package server

import (
	"fmt"
	"time"

	"log/slog"
	"net/http"

	"go.sakib.dev/le/logger"
	"go.sakib.dev/le/pkg/utils"
)

type Server struct {
	Dir     string
	Port    int
	state   ServerState
	eventCh chan ServerEventName
}

func NewServer(dir string, port int, ch chan ServerEventName) *Server {
	slog.SetDefault(slog.New(logger.NewHandler()))

	return &Server{
		Dir:     dir,
		Port:    port,
		eventCh: ch,
		state: ServerState{
			Conns: make(map[string]*Conn),
		},
	}
}

func (s *Server) Start() error {
	ch := make(chan ServerEvent, 100)
	handler := newHandler(s.Dir, ch)

	s.PrintUrl()

	go s.listenForData(ch)

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

	s.state.Addr = &url
	s.publish(EvNameAddrUpdated)

}

func (s *Server) publish(event ServerEventName) {
	if s.eventCh != nil {
		s.eventCh <- event
	}
}

func (s *Server) GetState() *ServerState {
	return &s.state
}

func (s *Server) listenForData(ch <-chan ServerEvent) {
	for data := range ch {
		switch data := data.(type) {
		case EventConnOpen:
			s.handleConnOpen(data)
		case EventConnClose:
			s.handleConnClose(data)
		case EventFileProgress:
			s.handleDownloadProgress(data)
		case EventDownloadStart:
			s.handleDownloadStart(data)
		default:
			slog.Warn("Unknown server event", "event", data)
		}
	}
}

func (s *Server) handleConnOpen(event EventConnOpen) {
	s.state.Conns[event.ConnID] = &Conn{
		ID:        event.ConnID,
		Client:    event.Client,
		UpdatedAt: event.Time,
		Filename:  event.Client.UserAgent, // Assuming filename is derived from UserAgent for simplicity
	}
	s.publish(EvNameConnOpen)
}

func (s *Server) handleConnClose(event EventConnClose) {
	if _, exists := s.state.Conns[event.ConnID]; exists {
		delete(s.state.Conns, event.ConnID)
		s.publish(EvNameConnClose)
	} else {
		slog.Warn("Connection close event for unknown connection", "conn_id", event.ConnID)
	}
}

func (s *Server) handleDownloadStart(event EventDownloadStart) {
	conn, exists := s.state.Conns[event.ConnID]
	if !exists {
		slog.Warn("Download start event for unknown connection", "conn_id", event.ConnID)
		return
	}

	conn.Filename = event.FileName
	conn.TotalSent = 0
	conn.UpdatedAt = event.Time

	s.publish(EvNameFileProgress)
}

func (s *Server) handleDownloadProgress(event EventFileProgress) {
	conn, exists := s.state.Conns[event.ConnID]
	if !exists {
		slog.Warn("File progress event for unknown connection", "conn_id", event.ConnID)
		return
	}

	conn.TotalSent += int64(event.Sent)
	conn.CurSpeed = int64(float64(event.Sent) / time.Since(conn.UpdatedAt).Seconds())
	conn.UpdatedAt = event.Time
	s.publish(EvNameFileProgress)
}
