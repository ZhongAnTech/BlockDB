package multiplexer

import (
	"encoding/hex"
	"fmt"
	"io"
)

type Observer interface {
	GetIncomingWriter() io.Writer
	GetOutgoingWriter() io.Writer
}

type ByteDumper struct {
	Name string
}

func (b *ByteDumper) Write(p []byte) (n int, err error) {
	fmt.Println(b.Name)
	fmt.Println(hex.Dump(p))
	return len(p), nil
}

type Dumper struct {
	incoming io.Writer
	outgoing io.Writer
}

func NewDumper(incomingName string, outgoingName string) *Dumper {
	return &Dumper{
		incoming: &ByteDumper{incomingName},
		outgoing: &ByteDumper{outgoingName},
	}
}
func (d *Dumper) GetIncomingWriter() io.Writer {
	return d.incoming
}

func (d *Dumper) GetOutgoingWriter() io.Writer {
	return d.outgoing
}
