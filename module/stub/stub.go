package stub

import (
	"crawler/log"
	"crawler/module"
)

// logger 代表日志记录器
var logger = log.DLogger()

// myModule 代表组件内部基础接口的实现类型
type myModule struct {
	// mid 代表组件ID
	mid module.MID
	// addr 代表组件的网络地址
	addr string
	// score 代表组件评分
	score uint64
	// scoreCalculator 代表评分计算器
	// scoreCalculator module
}
