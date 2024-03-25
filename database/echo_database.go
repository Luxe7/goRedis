package database

import (
	"goRedis/Interface/resp"
	"goRedis/resp/reply"
)

type EchoDatabase struct {
}

func NewDatabase() *EchoDatabase {
	return &EchoDatabase{}
}
func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) Close() {
	//TODO implement me
	panic("implement me")
}

func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	//TODO implement me
	panic("implement me")
}
