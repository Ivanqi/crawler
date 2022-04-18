package stub

import (
	"crawler/module"
	"testing"
)

// addStr 代表测试用的网络地址
var addStr = "127.0.0.1:8080"

// mid 代表测试用的MID
var mid = module.MID("D1|" + addStr)

func TestNew(t *testing.T) {
	mi, err := NewModuleInternal(mid, module.CalculateScoreSimple)
	if err != nil {
		t.Fatalf("创建内部模块时出错: %s(mid: %s)", err, mid)
	}

	if mi == nil {
		t.Fatal("无法创建内部模块!")
	}

	if mi.ID() != mid {
		t.Fatalf("内部模块的MID不一致。预期: %s, 实际: %s", mid, mi.ID())
	}

	if mi.Addr() != addStr {
		t.Fatalf("内部模块的地址不一致。预期: %s， 实际: %s", addStr, mi.Addr())
	}

	if mi.Score() != 0 {
		t.Fatalf("内部模块的score 不一致。预期: %d, 实际: %d", 0, mi.Score())
	}

	if mi.ScoreCalculator() == nil {
		t.Fatalf("内部模块的 score 计算器不一致. 预期: %p(%T), 实际: %p(%T)", module.CalculateScoreSimple, module.CalculateScoreSimple,
			mi.ScoreCalculator(), mi.ScoreCalculator())
	}

	if mi.CalledCount() != 0 {
		t.Fatalf("内部模块的调用计数不一致：预期: %d, 实际: %d", 0, mi.CalledCount())
	}

	if mi.AcceptedCount() != 0 {
		t.Fatalf("内部模块接受的计数不一致. 预期: %d， 实际: %d", 0, mi.AcceptedCount())
	}

	if mi.CompletedCount() != 0 {
		t.Fatalf("内部模块的完成计数不一致。预期: %d, 实际：%d", 0, mi.CompletedCount())
	}

	if mi.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理数不一致。预期: %d, 实际: %d", 0, mi.HandlingNumber())
	}

	illegalMID := module.MID("D127.0.0.1")
	mi, err = NewModuleInternal(illegalMID, nil)
	if err == nil {
		t.Fatalf("使用非法MID %q 创建内部模块时没有错误", illegalMID)
	}
}
