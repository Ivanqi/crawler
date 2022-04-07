package module

// Type 代表组件的类型
type Type string

// 当前认可的组件类型常量
const (
	// TYPE_DOWNLOADER 代表下载器
	TYPE_DOWNLOADER Type = "downloader"
	// TYPE_ANALYZER 代表分析器
	TYPE_ANALYZER Type = "analyzer"
	// TYPE_PIPELINE 代表条目处理管道
	TYPE_PIPELINE Type = "pipeline"
)

// legalTypeLetterMap 代表合法的组件类型-字母的映射
var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER:   "A",
	TYPE_PIPELINE:   "P",
}

// legalLetterTypeMap 代表合法的字母-组件类型的映射
var legalLetterTypeMap = map[string]Type{
	"D": TYPE_DOWNLOADER,
	"A": TYPE_ANALYZER,
	"P": TYPE_PIPELINE,
}

// CheckType 用于判断组件实例的类型是否匹配
func CheckType(moduleType Type, module )
