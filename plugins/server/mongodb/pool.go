package mongodb

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Pool struct {
	url      string
	max      int
	connChan chan net.Conn
	mu       sync.RWMutex
}

func NewPool(url string, maxConn int) *Pool {
	p := &Pool{}
	p.url = url
	p.max = maxConn
	return p
}

func (pool *Pool) newConn(url string) (net.Conn, error) {
	retrySleep := 50 * time.Millisecond
	for retryCount := 7; retryCount > 0; retryCount-- {
		c, err := net.Dial("tcp", url)
		if err == nil {
			return c, nil
		}
		// TODO log the error
		fmt.Println(fmt.Sprintf("dial error: %v", err))

		time.Sleep(retrySleep)
		retrySleep = retrySleep * 2
	}

	return nil, fmt.Errorf("failed to create new connection to url: %s", url)
}

func (pool *Pool) Release(c net.Conn) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.release(c)
}
func (pool *Pool) release(c net.Conn) error {
	pool.connChan <- c

	return nil
}

func (pool *Pool) Acquire() net.Conn {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conn := <-pool.connChan

	return conn
}
