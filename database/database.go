package database

import (
	"goRedis/Interface/resp"
	"goRedis/aof"
	"goRedis/config"
	"goRedis/lib/logger"
	"goRedis/resp/reply"
	"strconv"
	"strings"
)

type Database struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

func NewDatabase() *Database {
	num := config.Properties.Databases
	if num <= 0 {
		num = 16
	}
	database := &Database{dbSet: make([]*DB, num)}
	for i := range database.dbSet {
		tDb := NewDB()
		tDb.index = i
		database.dbSet[i] = tDb
	}
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAof(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
	}
	return database
}
func (db *Database) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	cmdName := strings.ToLower(string(args[0]))
	//数据库层面没有写select操作
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, db, args[1:])
	}
	dbIndex := client.GetDBIndex()
	nowDB := db.dbSet[dbIndex]
	return nowDB.Exec(client, args)
}

func (db *Database) Close() {
	//TODO implement me
	panic("implement me")
}

func (db *Database) AfterClientClose(c resp.Connection) {
	//TODO implement me
	panic("implement me")
}
func execSelect(c resp.Connection, database *Database, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(database.dbSet) {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
