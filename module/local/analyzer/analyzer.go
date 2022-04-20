package analyzer

import (
	log "crawler/logger"
	"crawler/module"
	"crawler/module/stub"
	"fmt"
)

// logger 代表日志记录器
var logger = log.DLogger()

// 分析器的实现类型
type myAnalyzer struct {
	// stub.ModuleInternal 代表组件基础实例
	stub.ModuleInternal
	// respParsers 代表响应解析器列表
	respParsers []module.ParseResponse
}

// New 用于创建一个分析器实例
func New(mid module.MID, respParsers []module.ParseResponse, scoreCalculator module.CalculateScore) (module.Analyzer, error) {
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil {
		return nil, err
	}

	if respParsers == nil {
		return nil, genParameterError("无响应解析器")
	}

	if len(respParsers) == 0 {
		return nil, genParameterError("空响应解析器列表")
	}

	var innerParsers []module.ParseResponse
	for i, parser := range respParsers {
		if parser == nil {
			return nil, genParameterError(fmt.Sprintf("无响应解析器[%d]", i))
		}
		innerParsers = append(innerParsers, parser)
	}

	return &myAnalyzer{
		ModuleInternal: moduleBase,
		respParsers:    innerParsers,
	}, nil
}
