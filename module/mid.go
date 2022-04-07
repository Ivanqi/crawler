package module

// DefaultSNGen 代表默认的组件序列号生成器。
var DefaultSNGen = NewSNGenertor(1, 0)

// midTemplate 代表组件ID的模板。
var midTemplate = "%s%d|%s"

// MID 代表组件ID。
type MID string

// GenMID 会根据给定参数生成组件ID
// func GenMID(mtype Type)
