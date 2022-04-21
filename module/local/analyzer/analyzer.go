package analyzer

import (
	log "crawler/logger"
	"crawler/module"
	"crawler/module/stub"
	"crawler/toolkit/reader"
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

func (analyzer *myAnalyzer) RespParsers() []module.ParseResponse {
	parser := make([]module.ParseResponse, len(analyzer.respParsers))
	copy(parser, analyzer.respParsers)
	return parser
}

func (analyzer *myAnalyzer) Analyze(resp *module.Response) (dataList []module.Data, errorList []error) {
	analyzer.ModuleInternal.IncrHandlingNumber()
	defer analyzer.ModuleInternal.DecrHandlingNumber()

	analyzer.ModuleInternal.IncrCalledCount()
	if resp == nil {
		errorList = append(errorList, genParameterError("无回应"))
		return
	}

	httpResp := resp.HTTPResp()
	if httpResp == nil {
		errorList = append(errorList, genParameterError("无 HTTP 响应"))
		return
	}

	httpReq := httpResp.Request
	if httpReq == nil {
		errorList = append(errorList, genParameterError("无 HTTP 请求"))
		return
	}

	var reqURL = httpReq.URL
	if reqURL == nil {
		errorList = append(errorList, genParameterError("无 HTTP 请求 URL"))
		return
	}

	analyzer.ModuleInternal.IncrAcceptedCount()
	respDepth := resp.Depth()
	logger.Infof("解析响应 (URL: %s, depth: %d)...\n", reqURL, respDepth)

	// 解析HTTP响应
	originalRespBody := httpResp.Body
	if originalRespBody != nil {
		defer originalRespBody.Close()
	}

	multipleReader, err := reader.NewMultipleReader(originalRespBody)
	if err != nil {
		errorList = append(errorList, genError(err.Error()))
		return
	}

	dataList = []module.Data{}
	for _, respParser := range analyzer.respParsers {
		httpResp.Body = multipleReader.Reader()
		pDataList, pErrorList := respParser(httpResp, respDepth)
		if pDataList != nil {
			for _, pData := range pDataList {
				if pData == nil {
					continue
				}
				dataList = appendDataList(dataList, pData, respDepth)
			}
		}

		if pErrorList != nil {
			for _, pError := range pErrorList {
				if pError == nil {
					continue
				}
				errorList = append(errorList, pError)
			}
		}
	}

	if len(errorList) == 0 {
		analyzer.ModuleInternal.IncrCompletedCount()
	}

	return dataList, errorList
}

// appendDataList 用于添加请求值或条目值到列表。
func appendDataList(dataList []module.Data, data module.Data, respDepth uint32) []module.Data {
	if data == nil {
		return dataList
	}

	req, ok := data.(*module.Request)
	if !ok {
		return append(dataList, data)
	}

	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = module.NewRequest(req.HTTPReq(), newDepth)
	}

	return append(dataList, req)
}
