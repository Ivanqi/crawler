package downloader

import (
	log "crawler/logger"
	"crawler/module"
	"crawler/module/stub"
	"net/http"
)

// logger 代表日志记录器
var logger = log.DLogger()

// myDownloader 代表下载器的实现类型。
type myDownloader struct {
	// stub.ModuleInternal 代表组件基础实例。
	stub.ModuleInternal
	// httpClient 代表下载用的HTTP客户端。
	httpClient http.Client
}

// New 用于创建一个下载器实例
func New(mid module.MID, client *http.Client, scoreCalculator module.CalculateScore) (module.Downloader, error) {
	moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, genParameterError("无 http 客户端")
	}

	return &myDownloader{
		ModuleInternal: moduleBase,
		httpClient:     *client,
	}, nil
}

func (downloader *myDownloader) Download(req *module.Request) (*module.Response, error) {
	downloader.ModuleInternal.IncrHandlingNumber()
	defer downloader.ModuleInternal.DecrHandlingNumber()

	downloader.ModuleInternal.IncrCalledCount()
	if req == nil {
		return nil, genParameterError("无请求")
	}

	httpReq := req.HTTPReq()
	if httpReq == nil {
		return nil, genParameterError("无http请求")
	}

	downloader.ModuleInternal.IncrAcceptedCount()
	logger.Infof("执行请求(URL: %s, depth: %d)... \n", httpReq.URL, req.Depth())
	httpResp, err := downloader.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	downloader.ModuleInternal.IncrCompletedCount()
	return module.NewResponse(httpResp, req.Depth()), nil
}
