package server

import "time"

type ServerEventName string

const (
	EvNameConnOpen      ServerEventName = "conn_open"
	EvNameConnClose     ServerEventName = "conn_close"
	EvNameDownloadStart ServerEventName = "download_start"
	EvNameFileProgress  ServerEventName = "file_progress"
	EvNameAddrUpdated   ServerEventName = "addr_updated"
)

type EventConnOpen struct {
	ConnID string
	Client *Client
	Time   time.Time
}

type Range struct {
	Start int64
	End   int64
}

type EventDownloadStart struct {
	ConnID    string
	FileName  string
	TotalSize int64
	Range     Range
	Time      time.Time
}

type EventConnClose struct {
	ConnID string
	Time   time.Time
}

type EventFileProgress struct {
	ConnID string
	Sent   int
	Time   time.Time
}

type ServerEvent interface {
	EventName() ServerEventName
}

func (e EventConnOpen) EventName() ServerEventName {
	return EvNameConnOpen
}
func (e EventConnClose) EventName() ServerEventName {
	return EvNameConnClose
}
func (e EventFileProgress) EventName() ServerEventName {
	return EvNameFileProgress
}
func (e EventDownloadStart) EventName() ServerEventName {
	return EvNameDownloadStart
}
