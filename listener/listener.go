package listener

import (
	"fmt"
	"github.com/annchain/BlockDB/processors"
	"github.com/sirupsen/logrus"
	"net"
)

type ProxyListener interface {
	Start(port int)
	Stop()
}
type ProxyListenerComponent struct {
	Listener ProxyListener
	Port     int
}

func NewProxyListenerComponent(l ProxyListener, port int) *ProxyListenerComponent {
	return &ProxyListenerComponent{
		Listener: l,
		Port:     port,
	}
}

func (l *ProxyListenerComponent) Start() {
	l.Listener.Start(l.Port)
}

func (l *ProxyListenerComponent) Stop() {
	l.Listener.Stop()
}

func (l *ProxyListenerComponent) Name() string {
	panic("implement me")
}

type GeneralTCPProxyListener struct {
	processor         processors.Processor
	port              int
	ln                net.Listener
	closed            bool
	maxConnectionSize int
}

func (l *GeneralTCPProxyListener) Name() string {
	return fmt.Sprintf("GeneralTCPProxyListener listening on %d", l.port)
}

func NewGeneralTCPListener(p processors.Processor, port int, maxConnectionSize int) *GeneralTCPProxyListener {
	return &GeneralTCPProxyListener{
		processor:         p,
		port:              port,
		maxConnectionSize: maxConnectionSize,
	}
}

func (l *GeneralTCPProxyListener) Start() {
	// start all prerequisites first. This is a block action
	// Do not return until ready. After this the listener will start to accept connections.
	l.processor.Start()

	go func() {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%v", l.port))
		if err != nil {
			logrus.WithError(err).WithField("port", l.port).Error("error listening on port")
			return
		}
		logrus.WithField("port", l.port).Info("server running on port")
		l.ln = ln
		// to limit the total number of accepted connections.
		maxChan := make(chan bool, l.maxConnectionSize)

		for {
			maxChan <- true
			conn, err := ln.Accept()
			if err != nil {
				if l.closed {
					break
				}
				logrus.WithError(err).Error("error accepting connection")
				continue
			}

			logrus.WithField("remote", conn.RemoteAddr()).Trace("accepted connection ")
			go func() {
				// release limit
				defer func(maxChan chan bool) { <-maxChan }(maxChan)
				l.processor.ProcessConnection(conn)
			}()
		}
	}()

}
func (l *GeneralTCPProxyListener) Stop() {
	l.closed = true
	err := l.ln.Close()
	if err != nil {
		logrus.WithError(err).Error("error closing connection")
	}
	l.processor.Stop()

}
