package database

import (
	"goRedis/Interface/resp"
	"goRedis/lib/wildcard"
	"goRedis/resp/reply"
)

func init() {
	RegisterCommand("del", execDel, -2)
	RegisterCommand("exists", execExists, -2)
	RegisterCommand("flushdb", execFlushDB, 1)
	RegisterCommand("type", execType, 2)
	RegisterCommand("rename", execRename, 3)
	RegisterCommand("renamenx", execRenamenx, 3)
	RegisterCommand("Keys", execKeys, 2)
}

// DEL
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	deleted := db.Removes(keys...)
	return reply.MakeIntReply(int64(deleted))
}

// EXISTS
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		if _, ok := db.GetEntity(string(arg)); ok {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// KEYS *
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0])) //取出通配符，判断匹配规则
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

// FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	return reply.MakeOkReply()
}

// TYPE
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	}
	return reply.MakeUnknownErrReply()
}

// RENAME
func execRename(db *DB, args [][]byte) resp.Reply {
	src, dest := string(args[0]), string(args[1])
	entity, exist := db.GetEntity(src)
	if exist != true {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, *entity)
	db.Remove(src)
	return reply.MakeOkReply()
}

// RENAMENX
func execRenamenx(db *DB, args [][]byte) resp.Reply {
	src, dest := string(args[0]), string(args[1])
	_, ok := db.GetEntity(src)
	if ok {
		return reply.MakeIntReply(0)
	}
	entity, exist := db.GetEntity(src)
	if exist != true {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, *entity)
	db.Remove(src)
	return reply.MakeIntReply(1)
}
