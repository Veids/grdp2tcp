package main

import (
	"log"

	"github.com/Veids/forwardlib/server/handler"
	"github.com/Veids/grdp2tcp/server/channel"
)

func main() {
	log.SetFlags(log.Ldate | log.Lshortfile)

	var err error
	c := channel.New("rdp2tcp")
	c.Init()
	err = c.Challenge()
	if err != nil {
		panic(err)
	}

	handler.Loop(&c)
}
