package tcp

import (
	"context"
	"goRedis/Interface/tcp"
	"goRedis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigchan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		logger.Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitGroup sync.WaitGroup
	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		waitGroup.Add(1)
		go func() {
			defer func() {
				waitGroup.Done()
			}()
			handler.Handler(ctx, conn)
		}()
	}
	waitGroup.Wait()
}
