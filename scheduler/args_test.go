package scheduler

import (
	"crawler/module"
	"crawler/module/local/analyzer"
	"crawler/module/local/downloader"
	"fmt"
	"net/http"
	"testing"
)

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

	requestArgs = genRequestArgs(nil, 0)
	if err := requestArgs.Check(); err == nil {
		t.Fatalf("检查结果不一致: 预期: %v, 实际: %v", nil, err)
	}

	// 测试Same方法的正确性
	one := genRequestArgs([]string{
		"bing.com",
	}, 0)

	another := genRequestArgs([]string{
		"bing.com",
	}, 0)

	same := one.Same(&another)
	if !same {
		t.Fatalf("不一致的请求参数相同性: 预期: %v, 实际: %v", true, same)
	}

	same = one.Same(nil)
	if same {
		t.Fatalf("与 nil 参数不一致的请求参数相同。预期: %v, 实际: %v", false, same)
	}

	another = genRequestArgs([]string{
		"bing.com",
	}, 1)

	same = one.Same(&another)

	if same {
		t.Fatalf("不同最大深度的不一致请求参数相同性。预期: %v, 实际: %v", false, same)
	}

	another = genRequestArgs(nil, 0)
	same = one.Same(&another)

	if same {
		t.Fatalf("不同接受域的请求参数相同性不一致：预期: %v, 实际: %v", false, same)
	}

	another = genRequestArgs([]string{
		"bing.net",
		"bing.com",
	}, 0)

	same = one.Same(&another)
	if same {
		t.Fatalf("不同接受域的请求参数相同性不一致: 预期: %v, 实际: %v", false, same)
	}

	one = genRequestArgs([]string{
		"sogou.com",
		"bing.com",
	}, 0)

	same = one.Same(&another)
	if same {
		t.Fatalf("不同接受域的请求参数相同性不一致. 预期: %v, 实际: %v", false, same)
	}
}

// genDataArgsByDetail 用于根据细致的参数生成数据参数的实例。
func genDataArgsByDetail(values [8]uint32) DataArgs {
	return DataArgs{
		ReqBufferCap:         values[0],
		ReqMaxBufferNumber:   values[1],
		RespBufferCap:        values[2],
		RespMaxBufferNumber:  values[3],
		ItemBufferCap:        values[4],
		ItemMaxBufferNumber:  values[5],
		ErrorBufferCap:       values[6],
		ErrorMaxBufferNumber: values[7],
	}
}

// genDataArgs 用于生成数据参数的实例。
func genDataArgs(bufferCap uint32, maxBufferNumber uint32, stepLen uint32) DataArgs {
	values := [8]uint32{}
	var bufferCapStep uint32
	var maxBufferNumberStep uint32
	for i := uint32(0); i < 8; i++ {
		if i%2 == 0 {
			values[i] = bufferCap + bufferCapStep*stepLen
			bufferCapStep++
		} else {
			values[i] = maxBufferNumber + maxBufferNumberStep*stepLen
			maxBufferNumberStep++
		}
	}
	return genDataArgsByDetail(values)
}

func TestArgsData(t *testing.T) {
	dataArgs := genDataArgs(10, 2, 1)
	if err := dataArgs.Check(); err != nil {
		t.Fatalf("检查结果不一致。预期: %v, 实际: %v", nil, err)
	}

	dataArgsList := []DataArgs{}
	for i := 0; i < 8; i++ {
		values := [8]uint32{2, 2, 2, 2, 2, 2, 2, 2}
		values[i] = 0
		dataArgsList = append(dataArgsList, genDataArgsByDetail(values))
	}

	for _, dataArgs := range dataArgsList {
		if err := dataArgs.Check(); err == nil {
			t.Fatalf("检查数据参数时没有错误! (dataArgs: %#v)", dataArgs)
		}
	}
}

// genSimpleDownloaders 用于生成一定数量的简易下载器
func genSimpleDownloaders(number int8, reuseMID bool, snGen module.SNGenertor, t *testing.T) []module.Downloader {
	if number < -1 {
		return []module.Downloader{nil}
	} else if number == -1 { // 不合规的MID
		mid := module.MID(fmt.Sprintf("A%d", snGen.Get()))
		httpClient := &http.Client{}
		d, err := downloader.New(mid, httpClient, nil)
		if err != nil {
			t.Fatalf("创建下载器时出错: %s(mid: %s, httpClient: %#v)", err, mid, httpClient)
		}
		return []module.Downloader{d}
	}

	results := make([]module.Downloader, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("D%d", snGen.Get()))
		}

		httpClient := &http.Client{}
		d, err := downloader.New(mid, httpClient, nil)
		if err != nil {
			t.Fatalf("创建下载器时出错: %s(mid: %s, httpClient:%#v)", err, mid, httpClient)
		}
		results[i] = d
	}

	return results
}

// genSimpleAnalyzers 用于生成一定数量的简易分析器
func genSimpleAnalyzers(number int8, resultMID bool, snGen module.SNGenertor, t *testing.T) []module.Analyzer {
	respParsers := []module.ParseResponse(parseATag)
	if number < -1 {
		return []module.Analyzer{nil}
	} else if number == -1 { // 不合规的MID
		mid := module.MID(fmt.Sprintf("P%d", snGen.Get()))
		a, err := analyzer.New(mid, respParsers, nil)
		if err != nil {
			t.Fatalf("创建分析器时出错: %s(mid: %s, respParsers: %#v)", err, mid, respParsers)
		}
		return []module.Analyzer{a}
	}

	results := make([]module.Analyzer, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !resultMID {
			mid = module.MID(fmt.Sprintf("A%d", snGen.Get()))
		}

		a, err := analyzer.New(mid, respParsers, nil)
		if err != nil {
			t.Fatalf("创建分析器时出错: %s (mid: %s, respParsers: %#v)", err, mid, respParsers)
		}
		results[i] = a
	}

	return results
}

// genSimpleModuleArgs 用于生成只包含简易组件实例的参数实例
func genSimpleModuleArgs(downloaderNumber int8, analyzerNumber int8, pipelineNumber int8, t *testing.T) ModuleArgs {
	snGen := module.NewSNGenertor(1, 0)

	return ModuleArgs{
		Downloaders: genSimpleDownloaders(downloaderNumber, false, snGen, t),
		Analyzers:   genSimpleAnalyzers(analyzerNumber, false, snGen, t),
		Pipelines:   genSimplePipelines(pipelineNumber, false, snGen, t),
	}
}

func TestArgsModule(t *testing.T) {
	moduleArgs := genSimpleModuleArgs
}
