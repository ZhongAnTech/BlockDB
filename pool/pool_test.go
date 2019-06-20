package pool

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"testing"
)

func TestPool(t *testing.T) {

	ln, err := net.Listen("tcp", fmt.Sprintf(":%v", 27017))
	if err != nil {
		logrus.WithError(err).WithField("port", 5656).Error("error listening on port")
		return
	}

	builder := NewDefaultTCPConnectionBuilder("172.28.152.101:27017")
	pool := NewSymmetricPool(builder)

	for {
		conn, err := ln.Accept()
		logrus.WithField("conn", conn.RemoteAddr()).Info("Accepted")
		if err != nil {
			logrus.WithError(err).Error("error accepting connection")
			return
		}
		go func() {
			// release limit
			targetConn, err := pool.MapConnection(conn)
			if err != nil {
				logrus.WithField("conn", conn.RemoteAddr()).WithError(err).Warn("error on connection")
			}
			pool.StartBidirectionalForwarding(conn, targetConn)
		}()
	}
}
