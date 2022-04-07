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

// Module 代表组件的基础接口
// 该接口的实现类型必须是并发安全
type Module interface {
	// ID 用于获取当前组件ID
	ID() MID
	// Addr 用于获取当前组件的网络地址的字符串形式
	Addr() string
	// Score 用于获取当前组件的评分
	Score() uint64
	// ScoreCalculator 用于获取评分计算器
	// ScoreCalculator() CalculateScore
	// CallCount 用于获取当前组件被调用的计数
	CalledCount() uint64
	// AcceptedCount
}
