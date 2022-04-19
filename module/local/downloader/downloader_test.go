package downloader

import (
	"bufio"
	"crawler/module"
	"crawler/module/stub"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	httpClient := &http.Client{}
	d, err := New(mid, httpClient, nil)
	if err != nil {
		t.Fatalf("创建下载器时出错: %s(mid: %s, httpClient: %#v)", err, mid, httpClient)
	}

	if d == nil {
		t.Fatal("无法创建下载器!")
	}

	if d.ID() != mid {
		t.Fatalf("下载器的MID 不一致。预期: %s, 实际: %s", mid, d.ID())
	}

	mid = module.MID("D127.0.0.1")
	d, err = New(mid, httpClient, nil)
	if err == nil {
		t.Fatalf("创建具有非法MID %q 的下载器时没有错误!", mid)
	}

	mid = module.MID("D1|127.0.0.1:8888")
	httpClient = nil
	d, err = New(mid, httpClient, nil)
	if err == nil {
		t.Fatal("使用 nil http 客户端创建下载器时没有错误")
	}
}

func TestDownload(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	httpClient := &http.Client{}

	d, _ := New(mid, httpClient, nil)
	url := "http://www.baidu.com/robots.txt"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("创建HTTP请求时出错: %s(url: %s)", err, url)
	}

	depth := uint32(0)
	req := module.NewRequest(httpReq, depth)
	resp, err := d.Download(req)
	if err != nil {
		t.Fatalf("下载内容时出错: %s(req: %#v)", err, req)
	}

	if resp == nil {
		t.Fatalf("无法为请求 %#v 创建下载", req)
	}

	if resp.Depth() != depth {
		t.Fatalf("不一致深度。预期: %d, 当前: %d", depth, resp.Depth())
	}

	httpResp := resp.HTTPResp()
	if httpResp == nil {
		t.Fatalf("无效的HTTP响应(url: %s)", url)
	}

	body := httpResp.Body
	if body == nil {
		t.Fatalf("HTTP响应正文无效!(url: %s)", url)
	}

	// NewReader 返回一个新的 Reader ，其缓冲区具有默认大小
	r := bufio.NewReader(body)
	line, _, err := r.ReadLine()
	if err != nil {
		t.Fatalf("读取HTTP响应正文时出错: %s(url: %s)", err, url)
	}

	lineStr := string(line)
	expectedFirstLine := "User-agent: Baiduspider"
	if lineStr != expectedFirstLine {
		t.Fatalf("HTTP响应正文的第一行不一致。预期: %s, 实际: %s(url: %s)", expectedFirstLine, lineStr, url)
	}

	// 测试参数有误的情况
	_, err = d.Download(nil)
	if err == nil {
		t.Fatal("使用nil请求下载时没有错误")
	}

	url = "http:///www.baidu.com/robots.txt"
	httpReq, err = http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("创建HTTP请求时出错: %s(url: %s)", err, url)
	}

	req = module.NewRequest(httpReq, 0)
	resp, err = d.Download(req)
	if err == nil {
		t.Fatalf("使用无效的url下载时没有错误%q!", url)
	}

	req = module.NewRequest(nil, 0)
	resp, err = d.Download(req)
	if err == nil {
		t.Fatal("使用 nil HTTP请求下载时没有错误")
	}
}

func TestCount(t *testing.T) {
	mid := module.MID("D1|127.0.0.1:8080")
	httpClient := &http.Client{}
	// 测试初始化后的计数
	d, _ := New(mid, httpClient, nil)
	di := d.(stub.ModuleInternal)
	if di.CalledCount() != 0 {
		t.Fatalf("内部模块的调用计数不一致。预期: %d, 实际: %d", 0, di.CalledCount())
	}

	if di.AcceptedCount() != 0 {
		t.Fatalf("内部模块接受的计数不一致。 预期:%d, 实际: %d", 0, di.AcceptedCount())
	}

	if di.CompletedCount() != 0 {
		t.Fatalf("内部模块的完成计数不一致。预期: %d, 实际: %d", 0, di.CompletedCount())
	}

	if di.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理数不一致。预期: %d, 实际: %d", 0, di.HandlingNumber())
	}

	// 测试处理失败时的计数
	d, _ = New(mid, httpClient, nil)
	di = d.(stub.ModuleInternal)
	url := "http:///www.baidu.com/robots.txt"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("创建HTTP请求时出错: %s(url: %s)", err, url)
	}

	req := module.NewRequest(httpReq, 0)
	_, err = d.Download(req)
	if di.CalledCount() != 1 {
		t.Fatalf("内部模块的调用计数不一致。预期:%d,实际: %d", 1, di.CalledCount())
	}

	if di.AcceptedCount() != 1 {
		t.Fatalf("内部模块接受的计数不一致。预期: %d, 实际: %d", 1, di.AcceptedCount())
	}

	if di.CompletedCount() != 0 {
		t.Fatalf("内部模块的完成计数不一致。 预期: %d, 实际: %d", 0, di.CompletedCount())
	}
	if di.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理编号不一致。预期: %d, 实际: %d", 0, di.HandlingNumber())
	}

	// 测试参数有误时的计数
	d, _ = New(mid, httpClient, nil)
	di = d.(stub.ModuleInternal)
	_, err = d.Download(nil)
	if di.CalledCount() != 1 {
		t.Fatalf("内部模块的调用计数不一致。预期: %d, 实际: %d", 1, di.CalledCount())
	}

	if di.AcceptedCount() != 0 {
		t.Fatalf("内部模块的调用计数不一致。预期: %d, 实际: %d", 0, di.AcceptedCount())
	}

	if di.CompletedCount() != 0 {
		t.Fatalf("内部模块的调用计数不一致。预期: %d, 实际: %d", 0, di.CompletedCount())
	}

	if di.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理数不一致。 预期: %d, 实际: %d", 0, di.HandlingNumber())
	}

	// 测试处理成功完成时的计数。
	d, _ = New(mid, httpClient, nil)
	di = d.(stub.ModuleInternal)
	url = "http://www.baidu.com/robots.txt"
	httpReq, err = http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("创建 HTTP 请求时出错: %s (url: %s)", err, url)
	}
	req = module.NewRequest(httpReq, 0)
	_, err = d.Download(req)
	if di.CalledCount() != 1 {
		t.Fatalf("内部模块接受的计数不一致: expected: %d, actual: %d", 1, di.CalledCount())
	}

	if di.AcceptedCount() != 1 {
		t.Fatalf("内部模块的完成计数不一致: expected: %d, actual: %d", 1, di.AcceptedCount())
	}

	if di.CompletedCount() != 1 {
		t.Fatalf("内部模块的处理编号不一致: expected: %d, actual: %d", 1, di.CompletedCount())
	}

	if di.HandlingNumber() != 0 {
		t.Fatalf("内部模块的处理编号不一致: expected: %d, actual: %d", 0, di.HandlingNumber())
	}
}
