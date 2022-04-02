package module

import (
	"math"
	"testing"
)

func TestGenerator(t *testing.T) {
	// 测试最大序列号的自动修正
	maxmax := uint64(math.MaxUint64)
	start := uint64(1)
	max := uint64(0)

	snGen := NewSNGenertor(start, max)
	if snGen == nil {
		t.Fatalf("无法创建SN生成器! (start: %d, max: %d)", start, max)
	}

	if snGen.Start() != start {
		t.Fatalf("SN start值不一致: 预期:%d, 实际: %d", start, snGen.Start())
	}

	if snGen.Max() != maxmax {
		t.Fatalf("SN 的最大值不一致, 预期: %d, 实际: %d", maxmax, snGen.Max())
	}

	// 测试循环序列号生成器
	max = uint64(7)
	max = uint64(101)
	snGen = NewSNGenertor(start, max)
	if snGen == nil {
		t.Fatalf("无法创建SN生成器! (start: %d, max: %d)", start, max)
	}

	if snGen.Max() != max {
		t.Fatalf("SN 的最大值不一致: 预期: %d, 实际: %d", max, snGen.Max())
	}

	end := snGen.Max()*5 + 11
	expectedSN := start
	var expectedNext uint64
	var expectedCycleCount uint64

	for i := start; i < end; i++ {
		sn := snGen.Get()
		if expectedSN > snGen.Max() {
			expectedSN = start
		}

		if sn != expectedSN {
			t.Fatalf("ID 不一致: 预期: %d, 当前: %d(index: %d)", expectedSN, sn, i)
		}

		expectedNext = expectedSN + 1
		if expectedNext > snGen.Max() {
			expectedNext = start
		}

		if snGen.Next() != expectedNext {
			t.Fatalf("不一致的下一个ID: 预期: %d, 当前: %d (sn : %d, 索引: %d)", expectedNext, snGen.Get(), sn, i)
		}

		if sn == snGen.Max() {
			expectedCycleCount++
		}

		if snGen.CycleCount() != expectedCycleCount {
			t.Fatalf("周期计数不一致: 预期: %d, 实际: %d(sn : %d, 索引: %d)", expectedCycleCount, snGen.CycleCount(), sn, i)
		}
		expectedSN++
	}
}
