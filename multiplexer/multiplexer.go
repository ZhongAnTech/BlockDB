package multiplexer

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type Multiplexer struct {
	observerFactory         ObserverFactory
	targetConnectionBuilder ConnectionBuilder
	closed                  bool
	biMapConn               *BiMapConn
}

func NewMultiplexer(targetConnectionBuilder ConnectionBuilder, observerFactory ObserverFactory) *Multiplexer {
	return &Multiplexer{
		observerFactory:         observerFactory,
		targetConnectionBuilder: targetConnectionBuilder,
		biMapConn:               NewBiMapConn(),
	}
}

func (m *Multiplexer) buildConnection() (target net.Conn, err error) {
	target, err = m.targetConnectionBuilder.BuildConnection()
	return
}

func (p *Multiplexer) StartBidirectionalForwarding(context DialogContext) {
	logrus.WithField("from", context.Source.RemoteAddr().String()).
		WithField("to", context.Target.RemoteAddr().String()).
		Info("start multiplexer bidirectional forwarding")

	observer := p.observerFactory.GetInstance(context)

	go p.keepForwarding(context.Source, context.Target, []*bufio.Writer{bufio.NewWriter(observer.GetIncomingWriter())})
	go p.keepForwarding(context.Target, context.Source, []*bufio.Writer{bufio.NewWriter(observer.GetOutgoingWriter())})
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

func (m *Multiplexer) ProcessConnection(source net.Conn) (err error) {
	// use connection builder to build a target connection
	// make pair
	// monitor them
	logrus.Trace("in")

	// build a writer
	target, err := m.buildConnection()
	if err != nil {
		logrus.WithError(err).Warn("failed to build target connection")
		// close the conn
		_ = source.Close()
		return err
	}
	// register both connection in the symmetric pool
	err = m.biMapConn.RegisterPair(source, target)
	if err != nil {
		_ = source.Close()
		_ = target.Close()
		return err
	}
	// build context to store connection info such as IP, identity, etc
	context := DialogContext{
		Source: source,
		Target: target,
	}

	m.StartBidirectionalForwarding(context)
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
	go func() {
		for !m.closed {
			logrus.WithField("size", m.biMapConn.Size()).Debug("poolsize")
			time.Sleep(time.Second * 60)
		}
	}()
}

func (m *Multiplexer) Stop() {
	m.closed = true
}
