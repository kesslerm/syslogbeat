package beater

import (
	"net"
	"time"

	"github.com/digitalocean/captainslog"
)

const (
	channelSize = 10
	bufferSize = 64 * 1024
)

type Server struct {
	udpAddr []string
	connections []net.PacketConn
	q chan captainslog.SyslogMsg
	shutdown bool
	location *time.Location
}

func NewServer(options ...func(*Server)) *Server {
	s := Server {
		q: make(chan captainslog.SyslogMsg, channelSize),
		location: time.UTC,
	}
	for _, option := range options {
		option(&s)
	}
	return &s
}

func OptionLocation(location *time.Location) func(*Server) {
	return func(s *Server) {
		s.location = location
	}
}

func OptionUDPAddr(addr string) func(*Server) {
	return func(s *Server) {
		s.udpAddr = append(s.udpAddr, addr)
	}
}

func (s *Server) listenUDP(addr string) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	c, err := net.ListenUDP("udp", a)
	if err != nil {
		return err
	}
	c.SetReadBuffer(bufferSize)

	go s.receiver(c)
	s.connections = append(s.connections, c)
	return nil
}

func (s *Server) Start() error {
	for _, addr := range s.udpAddr {
		err := s.listenUDP(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Stop() error {
	s.shutdown = true
	for _, c := range s.connections {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) receiver(c net.PacketConn) {
	buf := make([]byte, bufferSize)
    for {
		n, _, err := c.ReadFrom(buf)

		if err != nil {
			return
		}

		msg, err := captainslog.NewSyslogMsgFromBytes(buf[:n], captainslog.OptionLocation(s.location))
		if err == nil {
			s.q <- msg
		}
	}
}
