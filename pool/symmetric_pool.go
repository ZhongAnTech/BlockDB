package pool

import (
	"bufio"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

type SymmetricPool struct {
	MaxTargetConnectionSize int
	MaxSourceConnectionSize int
	builder                 ConnectionBuilder
	sourceTargetMap         map[net.Conn]net.Conn
	targetSourceMap         map[net.Conn]net.Conn
	lock                    sync.RWMutex
	closed                  bool
}

func NewSymmetricPool(builder ConnectionBuilder) *SymmetricPool {
	return &SymmetricPool{
		sourceTargetMap: make(map[net.Conn]net.Conn),
		targetSourceMap: make(map[net.Conn]net.Conn),
		builder:         builder,
	}
}

// constantly check if all connections are good, by select all of them.
func (p *SymmetricPool) relay(source net.Conn, target net.Conn, spectaculars []io.Writer) {
	logrus.Info("in")
	reader := bufio.NewReader(source)

	// writers
	var writers []*bufio.Writer
	writers = append(writers, bufio.NewWriter(target))
	for _, target := range spectaculars {
		writers = append(writers, bufio.NewWriter(target))
	}

	var buffer = make([]byte, 1024)
	for !p.closed {
		logrus.Info("gonna read bytes....")
		size, err := reader.Read(buffer)
		logrus.WithField("len", size).WithError(err).Info("read bytes")
		if err == io.EOF {
			logrus.Info("source closed")
			_ = p.Close(source)
			break
		} else if err != nil {
			logrus.WithError(err).Error("source error")
			_ = p.Close(source)
			return
		}
		//else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		//	logrus.Info("source timeout")
		//	conn.Close()
		//	return nil
		//}

		for _, writer := range writers {
			size2, err := writer.Write(buffer[0:size])
			_ = writer.Flush()
			logrus.WithError(err).WithField("len", size2).Info("wrote bytes")
			if err != nil {
				logrus.WithField("len", size2).WithError(err).Error("error on writing")
				_ = p.Close(source)
				break
			}
		}
	}
}

func (p *SymmetricPool) StartBidirectional(source net.Conn, target net.Conn) {
	logrus.Info("start bidirectional")
	go p.relay(source, target, []io.Writer{&Dumper{Name: "request"}})
	go p.relay(target, source, []io.Writer{&Dumper{Name: "response"}})
}

// MapConnection tries to build/reuse a backend connection according to the frontend connection
// If either way is closed, close the opposite one also.
func (p *SymmetricPool) MapConnection(source net.Conn) (targetConn net.Conn, err error) {
	if v, ok := p.sourceTargetMap[source]; ok {
		logrus.Info("connection reuse")
		return v, nil
	}
	logrus.Info("building conn")
	targetConn, err = p.builder.BuildConnection()
	if err != nil {
		logrus.WithError(err).Error("failed to built conn")
	}
	logrus.Info("built conn")

	// build a map
	p.lock.Lock()
	defer p.lock.Unlock()
	if v, tok := p.sourceTargetMap[source]; tok {
		// things changed after the first check, do not use this connection
		logrus.Info("built conn")
		err = targetConn.Close()
		return v, nil
	} else {
		logrus.Info("register pair")
		p.sourceTargetMap[source] = targetConn
		p.targetSourceMap[targetConn] = source
		return targetConn, nil
	}
	return targetConn, err
}

func (p *SymmetricPool) Close(source net.Conn) error {
	logrus.Info("closing both")
	err := source.Close()
	if err != nil {
		logrus.WithError(err).Warn("error on closing source")
	}
	if v, ok := p.sourceTargetMap[source]; ok {
		// close the backend
		err := v.Close()
		if err != nil {
			logrus.WithError(err).Warn("error on closing target")
		}

		delete(p.sourceTargetMap, source)
		delete(p.targetSourceMap, v)
		return err
	}
	if v, ok := p.targetSourceMap[source]; ok {
		// close the backend
		err := v.Close()
		if err != nil {
			logrus.WithError(err).Warn("error on closing target")
		}

		delete(p.targetSourceMap, source)
		delete(p.sourceTargetMap, v)
		return err
	}
	return nil
}
