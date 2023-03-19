package handlers

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/Veids/grdp2tcp/common"

	"github.com/hashicorp/yamux"
)

type ReverseServers struct {
	sync.RWMutex
	m map[string]*string
}

type ReverseHandler struct {
	dictionary ReverseServers
	session    *yamux.Session
}

func NewReverseHandler(session *yamux.Session) ReverseHandler {
	return ReverseHandler{
		ReverseServers{m: make(map[string]*string)},
		session,
	}
}

func (s *ReverseHandler) Serve() {
	for {
		stream, err := s.session.Accept()
		if err != nil {
			panic(err)
		}
		log.Println("Stream accepted")
		s.HandleStream(stream)
	}
}

func (s *ReverseHandler) HandleStream(stream net.Conn) {
	stype := make([]byte, 1)
	stream.Read(stype)

	switch stype[0] {
	case common.REVERSE_PORT_FORWARD:
		var addr common.Addr
		addr.Unmarshal(stream)
		remoteAddr := addr.ToString()

		s.dictionary.Lock()
		defer s.dictionary.Unlock()

		if val, ok := s.dictionary.m[remoteAddr]; ok {
			go s.Connect(stream, val)
		} else {
			log.Printf("Reverse remote address doesn't exists %s", remoteAddr)
			stream.Close()
		}
		break
	default:
		log.Printf("Invalid stream type %d", stype[0])
		stream.Close()
	}
}

func (s *ReverseHandler) Connect(stream net.Conn, localAddr *string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		log.Printf("Failed to resolve reverse addr: %v", err)
		stream.Close()
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("Failed to establish reverse tcp connection: %s", err)
		stream.Close()
	}

	go func() {
		log.Printf("Starting to copy conn to stream for %s", conn.RemoteAddr())
		io.Copy(conn, stream)
		conn.Close()
		log.Printf("Done copying conn to stream for %s", conn.RemoteAddr())
	}()

	go func() {
		log.Printf("Starting to copy stream to conn for %s", conn.RemoteAddr())
		io.Copy(stream, conn)
		stream.Close()
		log.Printf("Done copying stream to conn for %s", conn.RemoteAddr())
	}()
}
