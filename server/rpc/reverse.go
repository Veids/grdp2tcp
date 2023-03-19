package rpc

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/Veids/grdp2tcp/common"
	"github.com/hashicorp/yamux"
)

type ReverseServer struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup
	session  *yamux.Session
}

type ReverseServers struct {
	sync.RWMutex
	m map[string]*ReverseServer
}

func (s *ReverseServer) Serve(ip string, port uint32) {
	defer s.wg.Done()
	listenerAddr := s.listener.Addr().String()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("Error accepting a reverse client on %s", listenerAddr)
			}
		} else {
			if s.session == nil {
				log.Printf("Session on %s is nil", listenerAddr)
				conn.Close()
				panic(err)
			}

			log.Printf("[R:%s] Got client. Opening reverse stream for %s", listenerAddr, conn.RemoteAddr())

			stream, err := s.session.Open()
			if err != nil {
				log.Printf("[R:%s] Error opening reverse stream for %s: %v", listenerAddr, conn.RemoteAddr(), err)
				panic(err)
			}

			stream.Write([]byte{common.REVERSE_PORT_FORWARD})

			addr := common.Addr{
				Ip:   ip,
				Port: port,
			}
			addr.Marshal(stream)

			go func() {
				log.Printf("[R:%s] Starting to copy conn to stream for %s", listenerAddr, conn.RemoteAddr())
				io.Copy(conn, stream)
				conn.Close()
				log.Printf("[R:%s] Done copying conn to stream for %s", listenerAddr, conn.RemoteAddr())
			}()

			go func() {
				log.Printf("[R:%s] Starting to copy stream to conn for %s", listenerAddr, conn.RemoteAddr())
				io.Copy(stream, conn)
				stream.Close()
				log.Printf("[R:%s] Done copying stream to conn for %s", listenerAddr, conn.RemoteAddr())
			}()
		}
	}
}

func (s *ReverseServer) Stop() {
	log.Printf("Stopping listener 1")
	close(s.quit)
	log.Printf("Stopping listener 2")
	s.listener.Close()
	log.Printf("Stopping listener 3")
	s.wg.Wait()
	log.Printf("Stopping listener 4")
}
