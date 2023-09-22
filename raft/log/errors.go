package log

import "errors"

type ReplicateLogError error

var (
	// InconsistencyError AppendEntries RPC时用来告知`leader`，`follower`的日志不一致，需要被覆盖。
	InconsistencyError ReplicateLogError = errors.New("leader's logs inconsistency with leader")
)
