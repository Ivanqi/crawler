package downloader

import (
	log "crawler/logger"
	"crawler/module"
	"net/http"
)

// logger 代表日志记录器
var logger = log.DLogger()

// New 用于创建一个下载器实例
func New(mid module.MID, client *http.Client, scoreCalculator module.CalculateScore) (module.Downloader, error) {
	module.Base, err := stub.NewModuleInternal(mid, scoreCalculator)
}
