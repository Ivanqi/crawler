package pipeline

import (
	"crawler/errors"
	"testing"
)

func TestErrorGenError(t *testing.T) {
	simpleErrMsg := "testing error"
	expectedErrType := errors.ERROR_TYPE_PIPELINE

	err := genError(simpleErrMsg)
	ce, ok := err.(errors.CrawlerError)
	if !ok {
		t.Fatalf("错误类型不一致。预期: %T, 实际: %T", errors.NewCrawlerError("", ""), err)
	}

	if ce.Type() != expectedErrType {
		t.Fatalf("错误类型字符串不一致。预期: %q, 实际: %q", expectedErrType, ce.Type())
	}

	expectedErrMsg := "crawler error: pipeline error: " + simpleErrMsg
	if ce.Error() != expectedErrMsg {
		t.Fatalf("不一致的错误消息: 预期: %q, 实际: %q", expectedErrMsg, ce.Error())
	}
}
