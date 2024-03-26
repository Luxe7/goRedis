package database

import (
	"goRedis/Interface/resp"
	"goRedis/datastruct/dict"
	"goRedis/resp/reply"
	"strings"
)

type DB struct {
	index int
	data  dict.Dict
}
type ExecFunc func(db *DB, args [][]byte) resp.Reply
type CmdLine = [][]byte

func NewDB() *DB {
	db := &DB{data: dict.NewSyncDict()}
	return db
}
func (db *DB) Exec(connection resp.Connection, cmdline CmdLine) resp.Reply {
	cmdName := strings.ToLower(string(cmdline[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command" + cmdName)
	}
	if !validateArity(cmd.arity, cmdline) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.exector
	return fun(db, cmdline[1:])

}
func validateArity(arity int, cmdArgs [][]byte) bool {
	return true
}
