package main

import (
	"log"

	"github.com/Veids/grdp2tcp/server/channel"

	"github.com/hashicorp/yamux"
	"github.com/things-go/go-socks5"
)

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
		log.Println("Stream accepted")
		if err != nil {
			panic(err)
		}

		log.Println("Passing off to socks5")
		go func() {
			err := server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
