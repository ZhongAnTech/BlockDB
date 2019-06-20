package pool

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"time"
)

type SymmetricPool struct {
	MaxTargetConnectionSize int
	MaxSourceConnectionSize int
	biMapConn               *BiMapConn
	builder                 ConnectionBuilder
	closed                  bool
}

func NewSymmetricPool(builder ConnectionBuilder) *SymmetricPool {
	return &SymmetricPool{
		builder:   builder,
		biMapConn: NewBiMapConn(),
	}
}

// relay simply forward the content from source to target, without
func (p *SymmetricPool) relay(source net.Conn, target net.Conn, spectaculars []io.Writer) {
	logrus.Trace("in")
	reader := bufio.NewReader(source)

	// writers
	var writers []*bufio.Writer
	writers = append(writers, bufio.NewWriter(target))
	for _, target := range spectaculars {
		writers = append(writers, bufio.NewWriter(target))
	}

	var buffer = make([]byte, 1024)
	for !p.closed {
		logrus.Trace("gonna read bytes....")
		sizeRead, err := reader.Read(buffer)
		logrus.WithField("len", sizeRead).WithError(err).Trace("read bytes")
		if err == io.EOF {
			logrus.Info("source closed")
			_ = p.quitPair(source)
			break
		} else if err != nil {
			logrus.WithError(err).Trace("source error")
			_ = p.quitPair(source)
			break
		}

		// forward the message to all writers
		for _, writer := range writers {
			sizeWritten, err := writer.Write(buffer[0:sizeRead])
			_ = writer.Flush()
			logrus.WithError(err).WithField("len", sizeWritten).Trace("wrote bytes")
			if err != nil {
				logrus.WithField("len", sizeWritten).WithError(err).Trace("error on writing")
				break
			}
		}
	}
}

func (p *SymmetricPool) StartBidirectionalForwarding(source net.Conn, target net.Conn) {
	logrus.WithField("from", source.RemoteAddr().String()).WithField("to", target.RemoteAddr().String()).Info("start bidirectional")
	go p.relay(source, target, []io.Writer{&Dumper{Name: "request"}})
	go p.relay(target, source, []io.Writer{&Dumper{Name: "response"}})

	go func() {
		for {
			logrus.WithField("size", p.biMapConn.Size()).Info("poolsize")
			time.Sleep(time.Second)
		}
	}()

}

// MapConnection tries to build/reuse a backend connection according to the frontend connection
// If either way is closed, close the opposite one also.
func (p *SymmetricPool) MapConnection(source net.Conn) (targetConn net.Conn, err error) {
	logrus.Debug("building connection")
	targetConn, err = p.builder.BuildConnection()
	if err != nil {
		logrus.WithError(err).Error("failed to build conn")
		return
	}
	logrus.Debug("connection built")
	if targetConn == nil {
		panic("why nil")
	}

	err = p.biMapConn.RegisterPair(source, targetConn)
	if err != nil {
		_ = targetConn.Close()
	}
	return
}

func (p *SymmetricPool) quitPair(part net.Conn) (err error) {
	logrus.Debug("pair quitting")
	err = part.Close()
	if err != nil {
		logrus.WithError(err).Debug("error on closing part")
	}

	// find the counter part and close it also.
	counterPart := p.biMapConn.UnregisterPair(part)
	if counterPart == nil {
		// already unregistered by others
		return
	}
	err = counterPart.Close()
	if err != nil {
		logrus.WithError(err).Debug("error on closing counterpart")
	}

	return err
}
