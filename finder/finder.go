package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	lib "crawler/finder/internal"
	"crawler/finder/monitor"
	log "crawler/logger"
	sched "crawler/scheduler"
)

// 命令参数。
var (
	firstURL string
	domains  string
	depth    uint
	dirPath  string
)

// 日志记录器。
var logger = log.DLogger()

func init() {
	flag.StringVar(&firstURL, "first", "http://zhihu.sogou.com/zhihu?query=golang+logo", "The first URL which you want to access.")

	flag.StringVar(&domains, "domains", "zhihu.com",
		"The primary domains which you accepted. "+
			"Please using comma-separated multiple domains.")

	flag.UintVar(&depth, "depth", 3, "The depth for crawling.")

	flag.StringVar(&dirPath, "dir", "./pictures",
		"The path which you want to save the image files.")
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tfinder [flags] \n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()
	// 创建调度器。
	scheduler := sched.NewScheduler()
	// 准备调度器的初始化参数。
	domainParts := strings.Split(domains, ",")
	acceptedDomains := []string{}
	for _, domain := range domainParts {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			acceptedDomains = append(acceptedDomains, domain)
		}
	}

	requestArgs := sched.RequestArgs{
		AcceptedDomains: acceptedDomains,
		MaxDepth:        uint32(depth),
	}

	dataArgs := sched.DataArgs{
		ReqBufferCap:         50,   // 代表请求缓冲器的容量
		ReqMaxBufferNumber:   1000, // 代表请求缓冲器的最大数量
		RespBufferCap:        50,   // 代表响应缓冲器的容量
		RespMaxBufferNumber:  10,   // 代表响应缓冲器的最大数量
		ItemBufferCap:        50,   // 代表条目缓冲器的容量
		ItemMaxBufferNumber:  100,  // 代表条目缓冲器的最大数量
		ErrorBufferCap:       50,   // 代表错误缓冲器的容量
		ErrorMaxBufferNumber: 1,    // 代表错误缓冲器的最大数量
	}

	downloaders, err := lib.GetDownloaders(1)
	if err != nil {
		logger.Fatalf("创建下载程序时出错: %s", err)
	}

	analyzers, err := lib.GetAnalyzers(1)
	if err != nil {
		logger.Fatalf("创建分析器时出错: %s", err)
	}

	pipelines, err := lib.GetPipelines(1, dirPath)
	if err != nil {
		logger.Fatalf("创建管道时出错: %s", err)
	}

	moduleArgs := sched.ModuleArgs{
		Downloaders: downloaders,
		Analyzers:   analyzers,
		Pipelines:   pipelines,
	}

	// 初始化调度器。
	err = scheduler.Init(requestArgs, dataArgs, moduleArgs)
	if err != nil {
		logger.Fatalf("初始化计划程序时出错: %s", err)
	}

	// 准备监控参数。
	checkInterval := time.Second
	summarizeInterval := 100 * time.Millisecond
	maxIdleCount := uint(5)
	// 开始监控。
	checkCountChan := monitor.Monitor(scheduler, checkInterval, summarizeInterval, maxIdleCount, true, lib.Record)
	// 准备调度器的启动参数。
	firstHTTPReq, err := http.NewRequest("GET", firstURL, nil)
	if err != nil {
		logger.Fatalln(err)
		return
	}

	// 开启调度器
	err = scheduler.Start(firstHTTPReq)
	if err != nil {
		logger.Fatalf("启动计划程序时出错: %s", err)
	}

	// 等待监控结束。
	<-checkCountChan
}
