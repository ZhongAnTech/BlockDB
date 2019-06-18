package mongodb

import (
	"net"
	"sync"
)

type Pool struct {
	connChan chan net.Conn

	mu sync.RWMutex
}

func NewPool(maxConn int) *Pool {
	p := &Pool{}

	p.connChan = make(chan net.Conn, maxConn)

	return p
}

func (pool *Pool) newConn(url string) net.Conn {
	// TODO

	return nil
}

func (pool *Pool) FreeConn(c net.Conn) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.freeConn(c)
}
func (pool *Pool) freeConn(c net.Conn) error {
	pool.connChan <- c

	return nil
}

func (pool *Pool) Acquire() net.Conn {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conn := <-pool.connChan

	return conn
}
