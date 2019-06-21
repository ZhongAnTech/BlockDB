package multiplexer

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Multiplexer struct {
	source                  net.Conn
	target                  net.Conn
	observer                Observer
	targetConnectionBuilder ConnectionBuilder
	closed                  bool
	biMapConn               *BiMapConn
}

func NewMultiplexer(targetConnectionBuilder ConnectionBuilder, observers Observer) *Multiplexer {
	return &Multiplexer{
		observer:                observers,
		targetConnectionBuilder: targetConnectionBuilder,
		biMapConn:               NewBiMapConn(),
	}
}

func (m *Multiplexer) buildConnection() (target net.Conn, err error) {
	target, err = m.targetConnectionBuilder.BuildConnection()
	return
}

func (p *Multiplexer) StartBidirectionalForwarding() {
	logrus.WithField("from", p.source.RemoteAddr().String()).WithField("to", p.target.RemoteAddr().String()).Info("start multiplexer bidirectional forwarding")
	go p.keepForwarding(p.source, p.target, []*bufio.Writer{bufio.NewWriter(p.observer.GetIncomingWriter())})
	go p.keepForwarding(p.target, p.source, []*bufio.Writer{bufio.NewWriter(p.observer.GetOutgoingWriter())})

	go func() {
		for {
			logrus.WithField("size", p.biMapConn.Size()).Info("poolsize")
			time.Sleep(time.Second * 10)
		}
	}()
}

func (m *Multiplexer) keepForwarding(source net.Conn, target net.Conn, observerWriter []*bufio.Writer) {
	var buffer = make([]byte, 1024)
	reader := bufio.NewReader(source)
	writer := bufio.NewWriter(target)

	allWriters := []*bufio.Writer{writer}
	allWriters = append(allWriters, observerWriter...)

	for !m.closed {
		logrus.Trace("gonna read bytes....")
		sizeRead, err := reader.Read(buffer)
		logrus.WithField("len", sizeRead).WithError(err).Trace("read bytes")
		if err == io.EOF {
			logrus.WithField("addr", source.RemoteAddr().String()).Warn("EOF")
			_ = m.quitPair(source)
			break
		} else if err != nil {
			logrus.WithField("addr", source.RemoteAddr().String()).WithError(err).Warn("read error")
			_ = m.quitPair(source)
			break
		}

		// forward the message to all writers
		// TODO: optimize it so that observer do not block target
		for _, writer := range allWriters {
			sizeWritten, err := writer.Write(buffer[0:sizeRead])
			_ = writer.Flush()
			logrus.WithError(err).WithField("len", sizeWritten).Trace("wrote bytes")
			if err != nil {
				logrus.WithField("len", sizeWritten).WithError(err).Warn("error on writing")
				break
			}
		}
	}
}

func (m *Multiplexer) ProcessConnection(conn net.Conn) (err error) {
	// use connection builder to build a target connection
	// make pair
	// monitor them
	logrus.Trace("in")
	m.source = conn

	// build a writer
	m.target, err = m.buildConnection()
	if err != nil {
		logrus.WithError(err).Warn("failed to build target connection")
		// close the conn
		_ = conn.Close()
		return err
	}
	// register both connection in the symmetric pool
	err = m.biMapConn.RegisterPair(m.source, m.target)
	if err != nil {
		_ = m.source.Close()
		_ = m.target.Close()
		return err
	}
	m.StartBidirectionalForwarding()
	return nil
}

func (p *Multiplexer) quitPair(part net.Conn) (err error) {
	logrus.Debug("pair quitting")
	err = part.Close()
	if err != nil {
		logrus.WithError(err).Warn("error on closing part")
	}

	// find the counter part and close it also.
	counterPart := p.biMapConn.UnregisterPair(part)
	if counterPart == nil {
		// already unregistered by others
		return
	}
	err = counterPart.Close()
	if err != nil {
		logrus.WithError(err).Warn("error on closing counterpart")
	}

	return err
}

func (m *Multiplexer) Start() {
}

func (m *Multiplexer) Stop() {
	m.closed = true
}
