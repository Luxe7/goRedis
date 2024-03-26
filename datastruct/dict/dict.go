package dict

type Consume func(key string, val interface{}) bool

type Dict interface {
	Get(key string) (val interface{}, exists bool)
	len() int
	Put(key string, val interface{}) (result int)
	PutIfAbsent(key string, val interface{}) (result int)
	PutIfExists(key string, val interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consume Consume)
	keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	clear()
}
