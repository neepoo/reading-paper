package log

// Entry 表示一个logEntry
type Entry interface {
	// Command 返回需要作用于state machine的数据
	Command() []byte
	// Index 返回Entry在Log中的索引
	Index() uint64
	// Term 返回是在那个任期(term)内产生的Entry
	Term() uint64
}

// Storage 接口定义了logEntry在各个节点的操作
type Storage interface {
	// AppendEntry 新增一条log Entry, Follower节点的duplicate leader's logs 操作通过先DeleteEntry在AppendEntry来实现.
	AppendEntry(entry Entry) error
	// DeleteEntry 在指定索引删除其及其之后index的log entry
	DeleteEntry(at uint64) error
}
