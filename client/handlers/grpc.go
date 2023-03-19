package handlers

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/Veids/grdp2tcp/common"
	clientpb "github.com/Veids/grdp2tcp/protobuf/clientpb"
	commonpb "github.com/Veids/grdp2tcp/protobuf/commonpb"

	"github.com/hashicorp/yamux"
	"golang.org/x/net/context"
)

type ClientRpcServer struct {
	session *yamux.Session
	socks   SocksServers
	reverse *ReverseHandler
	control *rpc.Client
}

func NewClientRpcServer(session *yamux.Session, control *rpc.Client, reverse *ReverseHandler) *ClientRpcServer {
	return &ClientRpcServer{
		session,
		SocksServers{m: make(map[string]*SocksServer)},
		reverse,
		control,
	}
}

func (s *ClientRpcServer) SocksStart(ctx context.Context, addr *commonpb.Addr) (*commonpb.Empty, error) {
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

func (s *ClientRpcServer) SocksStop(ctx context.Context, addr *commonpb.Addr) (*commonpb.Empty, error) {
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

func (s *ClientRpcServer) ReverseStart(ctx context.Context, addrPack *commonpb.AddrPack) (*commonpb.Empty, error) {
	local_address := fmt.Sprintf("%s:%d", addrPack.Local.Ip, addrPack.Local.Port)
	remote_address := fmt.Sprintf("%s:%d", addrPack.Remote.Ip, addrPack.Remote.Port)

	s.reverse.dictionary.Lock()
	defer s.reverse.dictionary.Unlock()

	if _, ok := s.reverse.dictionary.m[remote_address]; ok {
		return nil, fmt.Errorf("Reverse listener %s already exist", remote_address)
	} else {
		s.reverse.dictionary.m[remote_address] = &local_address
		log.Printf("Added reverse handler %s", remote_address)
	}

	var reply string
	err := s.control.Call("ServerRpcServer.ReverseStart", common.Addr{
		Ip:   addrPack.Remote.Ip,
		Port: addrPack.Remote.Port,
	}, &reply)

	if err != nil {
		delete(s.reverse.dictionary.m, remote_address)
		return &commonpb.Empty{}, err
	}

	return &commonpb.Empty{}, nil
}

func (s *ClientRpcServer) ReverseStop(ctx context.Context, remoteAddr *commonpb.Addr) (*commonpb.Empty, error) {
	remote_address := fmt.Sprintf("%s:%d", remoteAddr.Ip, remoteAddr.Port)
	log.Printf("Trying to stop reverse %s", remote_address)

	s.reverse.dictionary.Lock()
	log.Printf("Trying to stop reverse %s. Lock", remote_address)
	defer s.reverse.dictionary.Unlock()
	log.Printf("Trying to stop reverse %s. Unlock", remote_address)

	if _, ok := s.reverse.dictionary.m[remote_address]; ok {
		var reply string
		log.Printf("Trying to stop reverse %s. Calling", remote_address)
		err := s.control.Call("ServerRpcServer.ReverseStop", common.Addr{
			Ip:   remoteAddr.Ip,
			Port: remoteAddr.Port,
		}, &reply)
		log.Printf("Trying to stop reverse %s. Call done", remote_address)
		if err != nil {
			log.Printf("Failed to stop reverse handler on the server: %s %v", reply, err)
		}

		delete(s.reverse.dictionary.m, remote_address)
		log.Printf("Removed reverse handler %s", remote_address)
	} else {
		return nil, fmt.Errorf("Reverse listener %s doesn't exist", remote_address)
	}

	return &commonpb.Empty{}, nil
}

func (s *ClientRpcServer) List(ctx context.Context, _ *commonpb.Empty) (*clientpb.EndpointList, error) {
	list := clientpb.EndpointList{}

	s.socks.RLock()
	for k := range s.socks.m {
		list.Endpoints = append(list.Endpoints, fmt.Sprintf("socks5 %s", k))
	}
	s.socks.RUnlock()

	s.reverse.dictionary.Lock()
	for k, v := range s.reverse.dictionary.m {
		list.Endpoints = append(list.Endpoints, fmt.Sprintf("reverse %s %s", *v, k))
	}
	s.reverse.dictionary.Unlock()

	return &list, nil
}
