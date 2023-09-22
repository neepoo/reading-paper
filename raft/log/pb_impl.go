package log

import gen "raft/proto/pb"

type ProtoLogEntry struct {
	gen.LogEntry
}

func (p *ProtoLogEntry) Command() []byte {
	return p.GetCommand()
}

func (p *ProtoLogEntry) Index() uint64 {
	return p.GetMeta().GetIndex()
}

func (p *ProtoLogEntry) Term() uint64 {
	return p.GetMeta().GetTerm()
}
