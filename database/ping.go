package database

import (
	"goRedis/Interface/resp"
	"goRedis/resp/reply"
)

func init() {
	RegisterCommand("ping", Ping, 1)
}

func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}
