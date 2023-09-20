package rpc

import (
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/Veids/grdp2tcp/common"

	"github.com/hashicorp/yamux"
)

type ServerRpcServer struct {
	session *yamux.Session
	reverse ReverseServers
}

var _ = reflect.TypeOf(ServerRpcServer{})

func NewServerRpcServer(session *yamux.Session) *ServerRpcServer {
	return &ServerRpcServer{
		session: session,
		reverse: ReverseServers{m: make(map[string]*ReverseServer)},
	}
}

func (s *ServerRpcServer) ReverseStart(addr *common.Addr, reply *string) error {
	address := fmt.Sprintf("%s:%d", addr.Ip, addr.Port)

	s.reverse.Lock()
	defer s.reverse.Unlock()

	if _, ok := s.reverse.m[address]; ok {
		return fmt.Errorf("Reverse listener %s already exist", address)
	} else {
		l, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}

		v := &ReverseServer{
			listener: l,
			quit:     make(chan interface{}),
			session:  s.session,
		}

		v.wg.Add(1)
		s.reverse.m[address] = v
		go v.Serve(addr.Ip, addr.Port)
		log.Printf("Started reverse server on %s:%d\n", addr.Ip, addr.Port)

		reply = nil
		return nil
	}
}

func (s *ServerRpcServer) ReverseStop(addr *common.Addr, reply *string) error {
	address := fmt.Sprintf("%s:%d", addr.Ip, addr.Port)
	s.reverse.Lock()
	defer s.reverse.Unlock()

	if val, ok := s.reverse.m[address]; ok {
		val.Stop()
		delete(s.reverse.m, address)
		log.Printf("Stopped reverse server on %s:%d\n", addr.Ip, addr.Port)
	} else {
		return fmt.Errorf("Reverse listener %s doesn't exist", address)
	}

	return nil
}
