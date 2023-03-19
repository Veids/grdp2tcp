package channel

import (
	"encoding/binary"
	"log"
	"os"
)

type channel struct {
	stdin     *os.File
	stdout    *os.File
	remaining []byte
}

func New() channel {
	return channel{stdin: os.Stdin, stdout: os.Stdout}
}

func (c *channel) Read(p []byte) (int, error) {
	var read uint32
	if len(c.remaining) == 0 {
		buf := make([]byte, 4)
		//TODO: Handle read size
		n, err := c.stdin.Read(buf)
		if n != 4 || err != nil {
			panic(err)
		}

		data_len := binary.LittleEndian.Uint32(buf)

		buf = make([]byte, data_len)
		read = 0
		for read < data_len {
			n, err = c.stdin.Read(buf[read:])
			if err != nil {
				panic(err)
			}
			read += uint32(n)
		}
		// log.Printf("Read %d bytes\n", read)
		c.remaining = buf
	}

	to_copy := len(c.remaining)
	if to_copy > len(p) {
		to_copy = len(p)
	}
	copy(p, c.remaining[:to_copy])

	if to_copy == len(c.remaining) {
		c.remaining = []byte{}
	} else {
		c.remaining = c.remaining[to_copy:]
	}

	return to_copy, nil
}

func (c *channel) Write(p []byte) (int, error) {
	return c.stdout.Write(p)
}

func (c *channel) Close() error {
	return nil
}

func (c *channel) Challenge() error {
	log.Printf("Waiting for a client\n")
	buf := make([]byte, 8)
	//TODO: handle read size
	c.Read(buf)

	//TODO: handle write size
	c.Write(buf)
	log.Printf("Client connected\n")
	return nil
}
