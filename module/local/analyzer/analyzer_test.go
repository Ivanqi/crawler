package analyzer

import (
	"bufio"
	"crawler/module"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// testingReader 代表测试专用的读取器，实现了io.ReadCloser接口类型
type testingReader struct {
	sr *strings.Reader
}

func (r testingReader) Read(b []byte) (n int, err error) {
	return r.sr.Read(b)
}

func (r testingReader) Close() error {
	return nil
}

// testingRespParser 为测试专用的响应解析函数
// 生成的函数会把响应的请求URL，响应体中的索引和响应深度存在条目中
func genTestingRespParser(fail bool) module.ParseResponse {
	if fail {
		return func(httpResp *http.Response, respDepth uint32) (data []module.Data, parseErrors []error) {
			errs := []error{fmt.Errorf("失败. (httpResp: %#v, respDepth: %#v)", httpResp, respDepth)}
			return nil, errs
		}
	}

	return func(httpResp *http.Response, respDepth uint32) (data []module.Data, parseErrors []error) {
		data = []module.Data{}
		parseErrors = []error{}
		item := module.Item(map[string]interface{}{})
		item["url"] = httpResp.Request.URL.String()
		bufReader := bufio.NewReader(httpResp.Body)
		line, _, err := bufReader.ReadLine()
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}

		lineStr := string(line)
		begin := strings.LastIndex(lineStr, "[")
		end := strings.LastIndex(lineStr, "]")

		if begin < 0 || end < 0 || begin > end {
			err := fmt.Errorf("wrong index for index: %d, %d", begin, end)
			parseErrors = append(parseErrors, err)
			return
		}

		index, err := strconv.Atoi(lineStr[begin+1 : end])
		if err != nil {
			parseErrors = append(parseErrors, err)
			return
		}

		item["index"] = index
		item["depth"] = respDepth
		data = append(data, item)
		req := module.NewRequest(nil, respDepth)
		data = append(data, req)
		return
	}
}

func TestNew(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	parsers := []module.ParseResponse{genTestingRespParser(false)}
	a, err := New(mid, parsers, nil)
	if err != nil {
		t.Fatalf("创建分析器时出错: %s(mid: %s)", err, mid)
	}

	if a == nil {
		t.Fatal("不能创建分析器!")
	}

	if a.ID() != mid {
		t.Fatalf("分析器的MID不一致。预期: %s，实际: %s", mid, a.ID())
	}

	if len(a.RespParsers()) != len(parsers) {
		t.Fatalf("管道的响应解析器编号不一致. 预期: %d, 实际: %d", len(a.RespParsers()), len(parsers))
	}

	// 测试参数有误的情况
	mid = module.MID("D127.0.0.1")
	a, err = New(mid, parsers, nil)
	if err == nil {
		t.Fatalf("创建具有非法MID %q 的分析器时没有错误!", mid)
	}

	mid = module.MID("D1|127.0.0.1:8080")
	parsersList := [][]module.ParseResponse{
		nil,
		[]module.ParseResponse{},
		[]module.ParseResponse{genTestingRespParser(false), nil},
	}

	for _, parsers := range parsersList {
		a, err = New(mid, parsers, nil)
		if err == nil {
			t.Fatalf("使用非法解析器: %#v 创建分析器时没有错误", parsers)
		}
	}
}
