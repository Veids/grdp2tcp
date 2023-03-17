package handlers

import (
	"fmt"
	"log"
	"net"

	clientpb "github.com/Veids/grdp2tcp/protobuf/clientpb"
	commonpb "github.com/Veids/grdp2tcp/protobuf/commonpb"

	"github.com/hashicorp/yamux"
	"golang.org/x/net/context"
)

type ClientRpcServer struct {
	session *yamux.Session
	socks   SocksServers
}

func NewClientRpcServer(session *yamux.Session) *ClientRpcServer {
	return &ClientRpcServer{session, SocksServers{m: make(map[string]*SocksServer)}}
}

func (s *ClientRpcServer) SocksStart(ctx context.Context, addr *clientpb.Addr) (*commonpb.Empty, error) {
	address := fmt.Sprintf("%s:%d", addr.Ip, addr.Port)
	s.socks.Lock()
	defer s.socks.Unlock()

	if _, ok := s.socks.m[address]; ok {
		return nil, fmt.Errorf("Socks listener %s already exist", address)
	} else {
		l, err := net.Listen("tcp", address)
		if err != nil {
			return nil, err
		}
		v := &SocksServer{
			listener: l,
			quit:     make(chan interface{}),
			session:  s.session,
		}
		v.wg.Add(1)
		s.socks.m[address] = v
		go v.Serve()
		log.Printf("Started socks server on %s:%d\n", addr.Ip, addr.Port)
		return &commonpb.Empty{}, nil
	}
}

func (s *ClientRpcServer) SocksStop(ctx context.Context, addr *clientpb.Addr) (*commonpb.Empty, error) {
	address := fmt.Sprintf("%s:%d", addr.Ip, addr.Port)
	s.socks.Lock()
	defer s.socks.Unlock()

	if val, ok := s.socks.m[address]; ok {
		val.Stop()
		delete(s.socks.m, address)
	} else {
		return nil, fmt.Errorf("Socks listener %s doesn't exist", address)
	}
	return &commonpb.Empty{}, nil
}

func (s *ClientRpcServer) ReverseStart(ctx context.Context, local_addr *clientpb.Addr, remote_addr *clientpb.Addr) (*commonpb.Empty, error) {
	return &commonpb.Empty{}, nil
}

func (s *ClientRpcServer) Stop(ctx context.Context, local_addr *clientpb.Addr, remote_addr *clientpb.Addr) (*commonpb.Empty, error) {
	return &commonpb.Empty{}, nil
}

func (s *ClientRpcServer) List(ctx context.Context, _ *commonpb.Empty) (*clientpb.EndpointList, error) {
	list := clientpb.EndpointList{}

	s.socks.RLock()
	defer s.socks.RUnlock()
	for k := range s.socks.m {
		list.Endpoints = append(list.Endpoints, fmt.Sprintf("socks5 %s", k))
	}

	return &list, nil
}
