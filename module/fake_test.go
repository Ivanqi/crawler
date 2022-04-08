package module

// defaultFakeDownloader 代表默认的仿造下载器
var defaultFakeDownloader = NewFakeDownloader(MID("D0"), CalculateScoreSimple)

// defaultFakeAnalyzer 代表默认的仿造分析器
var defaultFakeAnalyzer = NewFakeAnalyzer(MID("A1"), CalculateScoreSimple)

// defaultFakePipeline 代表默认的仿造条目处理管道
var defaultFakePipeline = NewFakePipeline(MID("P2"), CalculateScoreSimple)

// fakeModules 代表默认仿造组件的切片
var fakeModules = []Module{
	defaultFakeDownloader,
	defaultFakeAnalyzer,
	defaultFakePipeline,
}

// defaultFakeModuleMap 代表组件类型与默认仿造实例的映射
var defaultFakeModuleMap = map[Type]Module{
	TYPE_DOWNLOADER: defaultFakeDownloader,
	TYPE_ANALYZER:   defaultFakeAnalyzer,
	TYPE_PIPELINE:   defaultFakePipeline,
}

// fakeModuleFuncMap 代表组件类型与仿造实例生成函数的映射
var fakeModuleFuncMap = map[Type]func(mid MID) Module{
	TYPE_DOWNLOADER: func(mid MID) Module {
		return NewFakeDownloader(mid, CalculateScoreSimple)
	},
	TYPE_ANALYZER: func(mid MID) Module {
		return NewFakeAnalyzer(mid, CalculateScoreSimple)
	},
	TYPE_PIPELINE: func(mid MID) Module {
		return NewFakePipeline(mid, CalculateScoreSimple)
	},
}

// fakeModule 代表仿造的组件
type fakeModule struct {
	// mid 代表组件ID
	mid MID
	// score 代表组件评分
	score uint64
	// count 代表组件基础计数
	count uint64
	// scoreCalculator 代表评分计算器
	scoreCalculator CalculateScore
}

func (fm *fakeModule) ID() MID {
	return fm.mid
}
