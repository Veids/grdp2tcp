package main

import (
	"flag"
	"log"
	"os"

	"github.com/Veids/forwardlib/client/handler"
	"github.com/Veids/grdp2tcp/client/channel"
)

func main() {
	controlAddr := flag.String("c", "127.0.0.1:8337", "host:port")
	flag.Parse()
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Lshortfile)

	c := channel.New()
	c.Challenge()

	handler.Loop(&c, *controlAddr)
}
