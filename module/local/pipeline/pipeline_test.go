package pipeline

import (
	"crawler/module"
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
}
