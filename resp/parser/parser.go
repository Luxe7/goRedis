package parser

import (
	"bufio"
	"errors"
	"goRedis/Interface/resp"
	"goRedis/lib/logger"
	"goRedis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

type Payload struct {
	Data resp.Reply // it's not "Reply", but they are same
	Err  error
}

type readState struct {
	readingMultiline  bool
	expectedArgsCount int
	msgType           byte
	args              [][]byte
	bulkLen           int64
}

func (r *readState) finished() bool {
	return r.expectedArgsCount > 0 && r.expectedArgsCount == len(r.args)
}

func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parser0(reader, ch)
	return ch
}
func parser0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for true {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr {
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			ch <- &Payload{
				Err: errors.New("protocol error" + string(msg)),
			}
			state = readState{}
			continue
		}
		if !state.readingMultiline {
			if msg[0] == '*' {
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: err,
					}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
						Err:  nil,
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: err,
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 {
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
						Err:  nil,
					}
					state = readState{}
					continue
				}
			} else {
				res, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: res,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else {
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{Err: err}
				state = readState{}
				continue
			}
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  nil,
				}
				state = readState{}
			}
		}

	}

}

// readLine we can't use CRLF to split string but use the length of string
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 {
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error" + string(msg))
		}
	} else {
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error" + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine int64
	expectedLine, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error" + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.msgType = msg[0]
		state.readingMultiline = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error" + string(msg))
	}
}
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		v, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error" + string(msg))
		}
		result = reply.MakeIntReply(v)
	}
	return result, nil
}
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
