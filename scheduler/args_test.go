package scheduler

import (
	log "crawler/logger"
	"crawler/module"
	"crawler/module/local/analyzer"
	"crawler/module/local/downloader"
	"crawler/module/local/pipeline"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
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

// parseATag 代表一个响应解析函数的实现，只解析"A" 标签
func parseATag(httpResp *http.Response, respDepth uint32) ([]module.Data, []error) {
	// TODO: 支持更多的HTTP响应状态
	if httpResp.StatusCode != 200 {
		err := fmt.Errorf(fmt.Sprintf("不支持的状态码 %d! (httpResponse: %v)", httpResp.StatusCode, httpResp))
		return nil, []error{err}
	}

	reqURL := httpResp.Request.URL
	httpRespBody := httpResp.Body
	defer func() {
		if httpRespBody != nil {
			httpRespBody.Close()
		}
	}()

	var dataList []module.Data
	var errs []error
	// 开始解析
	doc, err := goquery.NewDocumentFromReader(httpRespBody)
	if err != nil {
		errs = append(errs, err)
		return dataList, errs
	}

	defer httpRespBody.Close()

	// 查找 'A' 标签并提取链接地址
	doc.Find("a").Each(func(index int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		// 前期过滤
		if !exists || href == "" || href == "#" || href == "/" {
			return
		}

		href = strings.TrimSpace(href)
		lowerHref := strings.ToLower(href)

		// 暂不支持对Javascript 代码的解析
		if href != "" && !strings.HasPrefix(lowerHref, "javascript") {
			aURL, err := url.Parse(href)
			if err != nil {
				log.DLogger().Warnf("解析标记 %q 中的属性 %q 时发生错误: %s (href: %s)", err, "href", "a", href)
				return
			}

			if !aURL.IsAbs() {
				aURL = reqURL.ResolveReference(aURL)
			}

			httpReq, err := http.NewRequest("GET", aURL.String(), nil)
			if err != nil {
				errs = append(errs, err)
			} else {
				req := module.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		}

		text := strings.TrimSpace(sel.Text())

		var id, name string
		if v, ok := sel.Attr("id"); ok {
			id = strings.TrimSpace(v)
		}
		if v, ok := sel.Attr("name"); ok {
			name = strings.TrimSpace(v)
		}

		m := make(map[string]interface{})
		m["a.parent"] = reqURL
		m["a.id"] = id
		m["a.name"] = name
		m["a.text"] = text
		m["a.index"] = index

		item := module.Item(m)
		dataList = append(dataList, item)
		log.DLogger().Infof("Processed item: %v", m)
	})

	return dataList, errs
}

// genSimpleAnalyzers 用于生成一定数量的简易分析器
func genSimpleAnalyzers(number int8, resultMID bool, snGen module.SNGenertor, t *testing.T) []module.Analyzer {
	respParsers := []module.ParseResponse{parseATag}
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
	moduleArgs := genSimpleModuleArgs(3, 2, 1, t)
	if err := moduleArgs.Check(); err != nil {
		t.Fatalf("Inconsistent check result: expected: %v, actual: %v",
			nil, err)
	}
	expectedSummary := ModuleArgsSummary{
		DownloaderListSize: 3,
		AnalyzerListSize:   2,
		PipelineListSize:   1,
	}
	summary := moduleArgs.Summary()
	if summary != expectedSummary {
		t.Fatalf("Inconsistent module args summary: expected: %#v, actual: %#v",
			expectedSummary, summary)
	}
	moduleArgsList := []ModuleArgs{
		genSimpleModuleArgs(0, 2, 1, t),
		genSimpleModuleArgs(3, 0, 1, t),
		genSimpleModuleArgs(3, 2, 0, t),
		ModuleArgs{},
	}
	for _, moduleArgs := range moduleArgsList {
		if err := moduleArgs.Check(); err == nil {
			t.Fatalf("No error when check module arguments! (moduleArgs: %#v)",
				moduleArgs)
		}
	}
}

// genSimplePipelines 用于生成一定数量的简易条目处理管道。
func genSimplePipelines(number int8, reuseMID bool, snGen module.SNGenertor, t *testing.T) []module.Pipeline {
	processors := []module.ProcessItem{processItem}
	if number < -1 {
		return []module.Pipeline{nil}
	} else if number == -1 { // 不合规的MID。
		mid := module.MID(fmt.Sprintf("D%d", snGen.Get()))
		p, err := pipeline.New(mid, processors, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)",
				err, mid, processors)
		}
		return []module.Pipeline{p}
	}
	results := make([]module.Pipeline, number)
	var mid module.MID
	for i := int8(0); i < number; i++ {
		if i == 0 || !reuseMID {
			mid = module.MID(fmt.Sprintf("P%d", snGen.Get()))
		}
		p, err := pipeline.New(mid, processors, nil)
		if err != nil {
			t.Fatalf("An error occurs when creating a pipeline: %s (mid: %s, processors: %#v)",
				err, mid, processors)
		}
		results[i] = p
	}
	return results
}

// processItem 代表一个条目处理函数的实现。
func processItem(item module.Item) (result module.Item, err error) {
	if item == nil {
		return nil, errors.New("Invalid item!")
	}
	// 生成结果。
	result = make(map[string]interface{})
	for k, v := range item {
		result[k] = v
	}
	if _, ok := result["number"]; !ok {
		result["number"] = len(result)
	}
	time.Sleep(10 * time.Millisecond)
	return result, nil
}
