package multiplexer

import (
	"bufio"
	"encoding/hex"
	"fmt"
)

type Observer interface {
	GetIncomingWriter() *bufio.Writer
	GetOutgoingWriter() *bufio.Writer
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
	incoming *bufio.Writer
	outgoing *bufio.Writer
}

func NewDumper(incomingName string, outgoingName string) *Dumper {
	return &Dumper{
		incoming: bufio.NewWriter(&ByteDumper{incomingName}),
		outgoing: bufio.NewWriter(&ByteDumper{outgoingName}),
	}
}
func (d *Dumper) GetIncomingWriter() *bufio.Writer {
	return d.incoming
}

func (d *Dumper) GetOutgoingWriter() *bufio.Writer {
	return d.outgoing
}
