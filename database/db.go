package database

import (
	"goRedis/Interface/database"
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

// 我们约定
// 对于不变长的命令，参数个数为正常记录
// 对于变长的命令，参数个数记录为带负号的最小值
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return arity == argNum
	}
	return argNum >= -arity
}

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if ok == false {
		return nil, false
	}
	entry, _ := raw.(*database.DataEntity)
	return entry, true
}

func (db *DB) PutEntity(key string, val *database.DataEntity) int {
	return db.data.Put(key, val)
}
func (db *DB) PutIfExists(key string, val *database.DataEntity) int {
	return db.data.PutIfExists(key, val)
}

func (db *DB) PutIfAbsent(key string, val *database.DataEntity) int {
	return db.data.PutIfAbsent(key, val)
}
func (db *DB) Remove(key string) int {
	return db.data.Remove(key)
}
func (db *DB) Removes(keys ...string) (deleted int) {
	for _, key := range keys {
		deleted += db.data.Remove(key)
	}
	return
}
func (db *DB) Flush() {
	db.data.Clear()
}
