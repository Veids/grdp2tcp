package handlers

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/Veids/grdp2tcp/common"

	"github.com/hashicorp/yamux"
)

type SocksServer struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	session  *yamux.Session
}

type SocksServers struct {
	sync.RWMutex
	m map[string]*SocksServer
}

func (s *SocksServer) Serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("Error accepting a client on %s", s.listener.Addr().String())
			}
		} else {
			if s.session == nil {
				log.Printf("Session on %s is nil", s.listener.Addr().String())
				conn.Close()
				panic(err)
			}
			log.Printf("Got client. Opening stream for %s", conn.RemoteAddr())

			stream, err := s.session.Open()
			if err != nil {
				log.Printf("Error opening stream for %s: %v", conn.RemoteAddr(), err)
				panic(err)
			}

			//TODO: Handle write size
			stream.Write([]byte{common.SOCKS})

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
	}
}

func (s *SocksServer) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}
