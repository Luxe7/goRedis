package reply

// PongReply is +Pong
type PongReply struct {
}

var pongBytes = []byte("+PONG\r\n")

// ToBytes marshal redis.Reply
func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

type OkReply struct {
}

var okBytes = []byte("+OK\r\n")

func (r *OkReply) ToBytes() []byte {
	return okBytes
}

// This saves memory because each time the MakeOkReply function is called, the same pointer is returned instead of creating a new OkReply object.
// This also avoids unnecessary memory allocation and garbage collection.
var theOkReply = new(OkReply)

// MakeOkReply returns an ok reply
func MakeOkReply() *OkReply {
	return theOkReply
}

var nullBulkBytes = []byte("$-1\r\n")

// NullBulkReply is empty string
type NullBulkReply struct{}

// ToBytes marshal redis.Reply
func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

// MakeNullBulkReply creates a new NullBulkReply
func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply is an empty list
type EmptyMultiBulkReply struct{}

// ToBytes marshal redis.Reply
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

// NoReply respond nothing, for commands like subscribe
type NoReply struct{}

var noBytes = []byte("")

// ToBytes marshal redis.Reply
func (r *NoReply) ToBytes() []byte {
	return noBytes
}
