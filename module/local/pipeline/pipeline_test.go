package pipeline

import (
	"crawler/module"
	"crawler/module/stub"
	"errors"
	"fmt"
	"testing"
)

func genTestingItemProccessor(fail bool) module.ProcessItem {
	if fail {
		return func(item module.Item) (result module.Item, err error) {
			return nil, fmt.Errorf("失败! (item: %#v)", item)
		}
	}

	return func(item module.Item) (result module.Item, err error) {
		num, ok := item["number"]
		if !ok {
			return nil, errors.New("无法寻找到number")
		}

		numInt, ok := num.(int)
		if !ok {
			return nil, fmt.Errorf("非整数 %v", num)
		}

		item["number"] = numInt + 1
		return item, nil
	}
}

func TestNew(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	processorNumber := 10
	processors := make([]module.ProcessItem, processorNumber)

	for i := 0; i < processorNumber; i++ {
		processors[i] = genTestingItemProccessor(false)
	}

	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错: %s(mid: %s, processors: %#v)", err, mid, processors)
	}

	if p == nil {
		t.Fatal("不能创建管道")
	}

	if p.ID() != mid {
		t.Fatalf("管道的MID不一致。预期: %s, 实际: %s", mid, p.ID())
	}

	if len(p.ItemProcessors()) != len(processors) {
		t.Fatalf("管道的项目处理器编号不一致。预期: %d, 实际: %d", len(p.ItemProcessors()), len(processors))
	}

	// 测试参数有误的情况
	mid = module.MID("D127.0.0.1")
	p, err = New(mid, processors, nil)
	if err == nil {
		t.Fatalf("创建具有非法MID %q 的管道时没有错误!", mid)
	}

	mid = module.MID("D1|127.0.0.1:8080")
	processors = append(processors, nil)
	p, err = New(mid, processors, nil)
	if err == nil {
		t.Fatal("使用 nil 处理器创建管道时没有错误")
	}

	processorsList := [][]module.ProcessItem{
		nil,
		[]module.ProcessItem{},
		[]module.ProcessItem{genTestingItemProccessor(false), nil},
	}

	for _, processors := range processorsList {
		p, err = New(mid, processors, nil)
		if err == nil {
			t.Fatal("使用 nil 处理器创建管道时没有错误")
		}
	}
}

func TestSend(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	processorNumber := 12
	processors := make([]module.ProcessItem, processorNumber)

	var expectedErrs int
	for i := 0; i < processorNumber; i++ {
		processors[i] = genTestingItemProccessor(false)
	}

	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错: %s(mid: %s, processors: %#v)", err, mid, processors)
	}

	errs := p.Send(nil)
	if len(errs) != 1 {
		t.Fatalf("Send() 后的错误编号不一致。预期: %d, 实际: %d", 1, len(errs))
	}

	item := module.Item(map[string]interface{}{"number": 0})
	errs = p.Send(item)
	number := item["number"].(int)
	if number != processorNumber {
		t.Fatalf("Send() 后项目中的数字不一致。预期: %d, 实际: %d", processorNumber, number)
	}

	if len(errs) != expectedErrs {
		t.Fatalf("Send() 后的错误编号不一致。预期: %d, 实际: %d", expectedErrs, len(errs))
	}

	// 测试存在处理失败的情况
	expectedErrs = 0
	for i := 0; i < processorNumber; i++ {
		if i%3 == 0 {
			processors[i] = genTestingItemProccessor(true)
			expectedErrs++
		} else {
			processors[i] = genTestingItemProccessor(false)
		}
	}

	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错。 %s(mid: %s, processors: %#v)", err, mid, processors)
	}

	item = module.Item(map[string]interface{}{"number": 0})
	errs = p.Send(item)
	if len(errs) != expectedErrs {
		t.Fatalf("Send() 后的错误编号不一致。预期: %d, 实际: %d", expectedErrs, len(errs))
	}

	// 测试把快递失败标记设置为true的情况
	p.SetFailFast(true)
	errs = p.Send(item)
	if len(errs) != 1 {
		t.Fatalf("Send() 后的错误编号不一致。预期: %d, 实际: %d", 1, len(errs))
	}

	// 测试把快速失败标记恢复为false 的情况
	p.SetFailFast(false)
	errs = p.Send(item)
	if len(errs) != expectedErrs {
		t.Fatalf("Send() 后的错误编号不一致。预期: %d, 实际: %d", expectedErrs, len(errs))
	}
}

func TestFailFast(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	processors := []module.ProcessItem{genTestingItemProccessor(false)}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错: %s(mid: %s, processors: %#v)", err, mid, processors)
	}

	if p.FailFast() != false {
		t.Fatalf("管道的失败快速标志不一致。预期: %v, 实际: %v", false, p.FailFast())
	}

	p.SetFailFast(true)
	if p.FailFast() != true {
		t.Fatalf("管道的失败快速标志不一致。预期: %v, 实际: %v", true, p.FailFast())
	}
}

func TestCount(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	processors := []module.ProcessItem{genTestingItemProccessor(false)}

	// 初始化后的计数
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错: %s(mid:%s, processors: %#v)", err, mid, processors)
	}

	pi, ok := p.(stub.ModuleInternal)
	if !ok {
		t.Fatal("无法将管道实例的类型转换为 stub.ModuleInternal")
	}

	if pi.CalledCount() != 0 {
		t.Fatalf("内部模块的调用计数不一致。预期: %d, 实际: %d", 0, pi.CalledCount())
	}
	if pi.AcceptedCount() != 0 {
		t.Fatalf("内部模块接受的计数不一致: 预期: %d, 实际: %d", 0, pi.AcceptedCount())
	}

	if pi.CompletedCount() != 0 {
		t.Fatalf("内部模块的完成计数不一致: 预期: %d, 实际: %d", 0, pi.CompletedCount())
	}

	if pi.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理编号不一致: 预期: %d, 实际: %d", 0, pi.HandlingNumber())
	}

	// 测试处理失败时的计数
	processors = []module.ProcessItem{genTestingItemProccessor(true)}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错: %s(mid: %s, processors: %#v)", err, mid, processors)
	}

	pi, ok = p.(stub.ModuleInternal)
	if !ok {
		t.Fatal("无法将管道实例的类型转换为 stub.ModuleInternal ")
	}

	item := module.Item(map[string]interface{}{"number": 0})
	p.Send(item)
	if pi.CalledCount() != 1 {
		t.Fatalf("内部模块的调用计数不一致: 预期: %d, 实际: %d", 1, pi.CalledCount())
	}

	if pi.AcceptedCount() != 1 {
		t.Fatalf("内部模块接受的计数不一致: 预期: %d, 实际: %d", 1, pi.AcceptedCount())
	}

	if pi.CompletedCount() != 0 {
		t.Fatalf("内部模块的完成计数不一致: 预期: %d, 实际: %d", 0, pi.CompletedCount())
	}

	if pi.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理编号不一致: 预期: %d, 实际: %d", 0, pi.HandlingNumber())
	}

	// 测试参数有误时的计数
	processors = []module.ProcessItem{genTestingItemProccessor(false)}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("创建管道时出错: %s. (mid: %s, processors: %#v)", err, mid, processors)
	}

	pi, ok = p.(stub.ModuleInternal)
	if !ok {
		t.Fatal("无法将管道实例的类型转换为 stub.ModuleInternal!")
	}

	p.Send(nil)
	if pi.CalledCount() != 1 {
		t.Fatalf("内部模块的调用计数不一致: 预期: %d, 实际: %d", 1, pi.CalledCount())
	}

	if pi.AcceptedCount() != 0 {
		t.Fatalf("内部模块接受的计数不一致: 预期: %d, 实际: %d", 0, pi.AcceptedCount())
	}

	if pi.CompletedCount() != 0 {
		t.Fatalf("内部模块的完成计数不一致: 预期: %d, 实际: %d", 0, pi.CompletedCount())
	}

	if pi.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理编号不一致: 预期: %d, 实际: %d", 0, pi.HandlingNumber())
	}

	// 测试处理成功完成时的计数
	processors = []module.ProcessItem{genTestingItemProccessor(false)}
	p, err = New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)", err, mid, processors)
	}

	pi, ok = p.(stub.ModuleInternal)
	if !ok {
		t.Fatal("Couldn't convert the type of pipeline instance to stub.ModuleInternal!")
	}

	p.Send(item)
	if pi.CalledCount() != 1 {
		t.Fatalf("Inconsistent called count for internal module: expected: %d, actual: %d",
			1, pi.CalledCount())
	}

	if pi.AcceptedCount() != 1 {
		t.Fatalf("Inconsistent accepted count for internal module: expected: %d, actual: %d", 1, pi.AcceptedCount())
	}

	if pi.CompletedCount() != 1 {
		t.Fatalf("Inconsistent completed count for internal module: expected: %d, actual: %d", 1, pi.CompletedCount())
	}

	if pi.HandlingNumber() != 0 {
		t.Fatalf("Inconsistent handling number for internal module: expected: %d, actual: %d", 0, pi.HandlingNumber())
	}
}

func TestSummary(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	processors := []module.ProcessItem{genTestingItemProccessor(false)}
	p, err := New(mid, processors, nil)
	if err != nil {
		t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)",
			err, mid, processors)
	}

	pi, ok := p.(*myPipeline)
	if !ok {
		t.Fatal("Couldn't convert the type of pipeline instance to stub.ModuleInternal!")
	}

	number := uint64(10000)
	for i := uint64(1); i < number; i++ {
		pi.IncrCalledCount()
		pi.IncrAcceptedCount()
		pi.IncrCompletedCount()
		pi.IncrHandlingNumber()
		if i%17 == 0 {
			pi.Clear()
		}
		counts := pi.Counts()
		expectedSummary := module.SummaryStruct{
			ID:        pi.ID(),
			Called:    counts.CalledCount,
			Accepted:  counts.AcceptedCount,
			Completed: counts.CompletedCount,
			Handling:  counts.HandlingNumber,
			Extra: extraSummaryStruct{
				FailFast:        pi.failFast,
				ProcessorNumber: len(pi.itemProcessors),
			},
		}
		summary := pi.Summary()
		if summary != expectedSummary {
			t.Fatalf("Inconsistent summary for internal module: expected: %#v, actual: %#v", expectedSummary, summary)
		}
	}
}
