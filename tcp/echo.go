package tcp

import (
	"bufio"
	"context"
	"goRedis/lib/logger"
	"goRedis/lib/sync/atomic"
	"goRedis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait //This waitGroup has time-out test
}

func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

type EchoHandle struct {
	activeConn sync.Map
	closing    atomic.Boolean //atomic.Boolean to avoid the risk of concurrency
}

func NewEchoHandler() *EchoHandle {
	return &EchoHandle{}

}

func (handler *EchoHandle) Handler(ctx context.Context, conn net.Conn) {
	if handler.closing.Get() {
		_ = conn.Close()
	}
	client := &EchoClient{Conn: conn}
	handler.activeConn.Store(client, struct {
	}{})
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connect close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (handler *EchoHandle) Close() error {
	logger.Info("handler shutting down")
	handler.closing.Set(true)
	handler.activeConn.Range(func(key, value any) bool {
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		return true

	})
	return nil
}
