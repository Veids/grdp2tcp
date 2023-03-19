package main

import (
	"log"
	"net"
	"net/rpc"
	"os"

	"github.com/Veids/grdp2tcp/client/channel"
	"github.com/Veids/grdp2tcp/client/handlers"
	"github.com/Veids/grdp2tcp/common"
	"github.com/Veids/grdp2tcp/protobuf/clientpb"

	"github.com/hashicorp/yamux"
	"google.golang.org/grpc"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Lshortfile)

	c := channel.New()
	c.Challenge()

	session, err := yamux.Client(&c, nil)
	if err != nil {
		log.Printf("Error creating client in yamux")
		panic(err)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:8337")
	if err != nil {
		panic(err)
	}

	cstream, err := session.Open()
	cstream.Write([]byte{common.CONTROL})
	if err != nil {
		panic(err)
	}
	control := rpc.NewClient(cstream)
	reverse := handlers.NewReverseHandler(session)
	go reverse.Serve()

	s := handlers.NewClientRpcServer(session, control, &reverse)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	clientpb.RegisterClientRpcServer(grpcServer, s)
	grpcServer.Serve(lis)
}
