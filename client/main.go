package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"client/channel"

	"github.com/hashicorp/yamux"
)

func listenForClients(listen string, port int, session *yamux.Session) error {
	var ln net.Listener
	var address string
	var err error

	address = fmt.Sprintf("%s:%d", listen, port)
	log.Printf("Listening on %s\n", address)
	ln, err = net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting on %s: %v", address, err)
			panic(err)
		}

		if session == nil {
			log.Printf("Session on %s is nil", address)
			conn.Close()
			panic(err)
		}
		log.Printf("Got client. Opening stream for %s", conn.RemoteAddr())

		stream, err := session.Open()
		if err != nil {
			log.Printf("Error opening stream for %s: %v", conn.RemoteAddr(), err)
			panic(err)
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
}

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

	listenForClients("127.0.0.1", 1080, session)
}
