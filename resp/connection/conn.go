package connection

import (
	"net"
	"sync"
	"time"

	"goRedis/lib/sync/wait"
)

type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait
	mu           sync.Mutex
	selectDB     int
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn:         conn,
		waitingReply: wait.Wait{},
		mu:           sync.Mutex{},
		selectDB:     0,
	}
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

func (c *Connection) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()
	_, err := c.conn.Write(bytes)
	return err
}

func (c *Connection) GetDBIndex() int {
	return c.selectDB
}

func (c *Connection) SelectDB(dbNum int) {
	c.selectDB = dbNum
}
