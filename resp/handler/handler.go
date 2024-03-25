package handler

import (
	"context"
	"goRedis/database"
	"goRedis/resp/parser"
	"goRedis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"

	databaseface "goRedis/Interface/database"
	"goRedis/lib/logger"
	"goRedis/lib/sync/atomic"
	"goRedis/resp/connection"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         databaseface.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databaseface.Database
	db = database.NewDatabase()
	return &RespHandler{
		db: db,
	}
}

func (r RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}

func (r *RespHandler) Handler(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		_ = conn.Close()
	}
	client := connection.NewConn(conn)
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(conn)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "user of closed network connection") {
				r.closeClient(client)
				logger.Info("connection closed" + client.RemoteAddr().String())
				return
			}
			//protocol error
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("connection closed" + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			continue
		}
		reply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := r.db.Exec(client, reply.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}

	}

}

func (r *RespHandler) Close() error {
	logger.Info("Handler is shutting down")
	r.closing.Set(true)
	r.activeConn.Range(
		func(key, value interface{}) bool {
			client := key.(*connection.Connection)
			_ = client.Close()
			return true
		},
	)
	r.db.Close()
	return nil
}
