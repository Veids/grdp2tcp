package main

import (
	"log"
	"net"
	"net/rpc"
	"sync"

	"github.com/Veids/grdp2tcp/common"
	"github.com/Veids/grdp2tcp/server/channel"
	rrpc "github.com/Veids/grdp2tcp/server/rpc"

	"github.com/hashicorp/yamux"
	"github.com/things-go/go-socks5"
)

type Control struct {
	sync.RWMutex
	stream          *net.Conn
	serverRpcServer *rrpc.ServerRpcServer
}

type Handler struct {
	server  *socks5.Server
	control Control
	session *yamux.Session
}

func (s *Handler) HandleStream(stream net.Conn) {
	//TODO: Handle read size
	stype := make([]byte, 1)
	stream.Read(stype)

	switch stype[0] {
	case common.CONTROL:
		s.control.Lock()
		if s.control.stream != nil {
			log.Println("Control stream already defined")
		} else {
			s.control.stream = &stream
			s.control.serverRpcServer = rrpc.NewServerRpcServer(s.session)
			r := rpc.NewServer()
			r.Register(s.control.serverRpcServer)
			go r.ServeConn(stream)
		}
		s.control.Unlock()
		break
	case common.SOCKS:
		log.Println("Passing off to socks5")
		go func() {
			err := s.server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
		break
	default:
		log.Printf("Invalid stream type %d", stype[0])
		stream.Close()
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Lshortfile)

	var err error
	var session *yamux.Session
	c := channel.New("rdp2tcp")
	c.Init()
	err = c.Challenge()
	if err != nil {
		panic(err)
	}

	session, err = yamux.Server(&c, nil)
	handler := Handler{
		server:  socks5.NewServer(),
		session: session,
	}

	for {
		stream, err := session.Accept()
		if err != nil {
			panic(err)
		}
		log.Println("Stream accepted")
		handler.HandleStream(stream)
	}
}
