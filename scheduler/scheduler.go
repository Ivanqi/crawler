package scheduler

// logger 代表日志记录器

// Scheduler 代表调度器的接口类型
type Scheduler interface {
	// Init 用于初始化调度器, 参数requestArgs代表请求相关的参数, 参数dataArgs代表数据相关的参数, 参数moduleArgs代表组件相关的参数
	Init(requestArgs int)
}
