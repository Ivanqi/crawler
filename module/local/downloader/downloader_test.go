package downloader

import (
	"crawler/module"
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
