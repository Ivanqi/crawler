package downloader

import (
	"crawler/errors"
	"testing"
)

func TestErrorGenError(t *testing.T) {
	simpleErrMsg := "testing error"
	expectedErrType := errors.ERROR_TYPE_DOWNLOADER
	err := genError(simpleErrMsg)
	ce, ok := err.(errors.CrawlerError)
	if !ok {
		t.Fatalf("非法错误类型. 预期: %T, 实际: %T", errors.NewCrawlerError("", ""), err)
	}

	if ce.Type() != expectedErrType {
		t.Fatalf("非法错误类型字符. 预期: %q, 实际: %q", expectedErrType, ce.Type())
	}

	expectedErrMsg := "crawler error: downloader error: " + simpleErrMsg
	if ce.Error() != expectedErrMsg {
		t.Fatalf("非法错误类型信息: 预期: %q, 实际: %q", expectedErrMsg, ce.Error())
	}
}
