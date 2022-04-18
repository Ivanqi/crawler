package scheduler

import "testing"

// genRequestArgs 用于生成请求参数的实例。
func genRequestArgs(acceptedDomains []string, maxDepth uint32) RequestArgs {
	return RequestArgs{
		AcceptedDomains: acceptedDomains,
		MaxDepth:        maxDepth,
	}
}

func TestArgsRequest(t *testing.T) {
	requestArgs := genRequestArgs([]string{}, 0)
	if err := requestArgs.Check(); err != nil {
		t.Fatalf("检查结果不一致。 预期: %v, 实际: %v", nil, err)
	}
}
