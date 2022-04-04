package module

// Counts 代表用于汇集组件内部计数的类型
type Counts struct {
	// CalledCount 代表调用计数
	CalledCount uint64
	// AcceptedCount 代表接受计数
	AcceptedCount uint64
	// CompletedCount 代表成功完成计数
	CompletedCount uint64
	// HandlingNumber 代表实时处理数
	HandlingNumber uint64
}

// SummaryStruct 代表组件摘要结构的类型
type SummaryStruct struct {
	ID        MID         `json:"id"`
	Called    uint64      `json:"called"`
	Accepted  uint64      `json:"accepted"`
	Completed uint64      `json:"completed"`
	Handling  uint64      `json:"handling"`
	Extra     interface{} `json:"extra,omitempty"`
}
