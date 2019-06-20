package pool

import (
	"errors"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type BiMapConn struct {
	sourceTargetMap map[net.Conn]net.Conn
	targetSourceMap map[net.Conn]net.Conn
	mapLock         sync.RWMutex
}

func NewBiMapConn() *BiMapConn {
	return &BiMapConn{
		sourceTargetMap: make(map[net.Conn]net.Conn),
		targetSourceMap: make(map[net.Conn]net.Conn),
	}
}

func (b *BiMapConn) RegisterPair(source net.Conn, target net.Conn) error {
	b.mapLock.Lock()
	defer b.mapLock.Unlock()

	if _, ok := b.sourceTargetMap[source]; ok {
		return errors.New("duplicate source")
	}
	if _, ok := b.targetSourceMap[target]; ok {
		return errors.New("duplicate target")
	}
	b.sourceTargetMap[source] = target
	b.targetSourceMap[target] = source
	return nil
}

func (b *BiMapConn) UnregisterPair(part net.Conn) (counterPart net.Conn) {
	b.mapLock.Lock()
	defer b.mapLock.Unlock()

	logrus.WithField("part", part.RemoteAddr().String()).Info("unregistering")

	if v, ok := b.sourceTargetMap[part]; ok {
		counterPart = v
		delete(b.sourceTargetMap, part)
		delete(b.targetSourceMap, v)
		return
	}
	if v, ok := b.targetSourceMap[part]; ok {
		counterPart = v
		delete(b.targetSourceMap, part)
		delete(b.sourceTargetMap, v)
		return
	}
	return nil
}

func (b *BiMapConn) GetCounterPart(part net.Conn) (counterPart net.Conn) {
	if v, ok := b.sourceTargetMap[part]; ok {
		counterPart = v
		return
	}
	if v, ok := b.targetSourceMap[part]; ok {
		counterPart = v
		return
	}
	return nil
}

func (b *BiMapConn) Size() int {
	return len(b.sourceTargetMap)
}
