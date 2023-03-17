package main

import (
	"log"
	"net"

	"github.com/Veids/grdp2tcp/server/channel"

	"github.com/hashicorp/yamux"
	"github.com/things-go/go-socks5"
)

const (
	SOCKS byte = iota
)

func handleStream(server *socks5.Server, stream net.Conn) {
	//TODO: Handle read size
	stype := make([]byte, 1)
	stream.Read(stype)

	switch stype[0] {
	case SOCKS:
		log.Println("Passing off to socks5")
		go func() {
			err := server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
	default:
		log.Printf("Invalid stream type %d", stype[0])
		stream.Close()
	}
}

func main() {
	var err error
	var session *yamux.Session
	c := channel.New("rdp2tcp")
	c.Init()
	err = c.Challenge()
	if err != nil {
		panic(err)
	}

	session, err = yamux.Server(&c, nil)
	server := socks5.NewServer()

	for {
		stream, err := session.Accept()
		if err != nil {
			panic(err)
		}
		log.Println("Stream accepted")
		handleStream(server, stream)
	}
}
