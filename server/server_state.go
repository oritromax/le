package server

import (
	"time"
)

type Client struct {
	IP          string
	Host        string
	UserAgent   string
	ConnectedAt time.Time
}

type Conn struct {
	ID        string
	Client    *Client
	TotalSent int64
	CurSpeed  int64
	UpdatedAt time.Time
	Filename  string
}

type ServerState struct {
	Dir   string
	Addr  *string
	Conns map[string]*Conn
}
