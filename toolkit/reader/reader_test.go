package reader

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReaderNew(t *testing.T) {
	expectedData := "0987dcba"
	rr, err := NewMultipleReader(strings.NewReader(expectedData))
	if err != nil {
		t.Fatalf("新建多重读取器发生错误: %s", err)
	}

	buffer := new(bytes.Buffer)
	// 把rr.Reader() 复制到buffer
	_, err = io.Copy(buffer, rr.Reader())
	if err != nil {
		t.Fatalf("复制数据时出错: %s", err)
	}

	content1 := buffer.String()
	if content1 != expectedData {
		t.Fatalf("不一致的数据: 预期: %s, 实际: %s", expectedData, content1)
	}

	content2 := buffer.String()
	if content2 != expectedData {
		t.Fatalf("不一致的数据: 预期: %s, 实际: %s", expectedData, content2)
	}
}
