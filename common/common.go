package common

import (
	"encoding/gob"
	"fmt"
	"io"
)

type Addr struct {
	Ip   string
	Port uint32
}

func (s *Addr) Marshal(wire io.Writer) error {
	enc := gob.NewEncoder(wire)

	return enc.Encode(s)
}

func (s *Addr) Unmarshal(wire io.Reader) error {
	dec := gob.NewDecoder(wire)

	return dec.Decode(s)
}

func (s *Addr) ToString() string {
	return fmt.Sprintf("%s:%d", s.Ip, s.Port)
}

const (
	CONTROL byte = iota
	SOCKS
	PORT_FORWARD
	REVERSE_PORT_FORWARD
)
