package module

// CalculateScore 代表用于计算组件评分的函数类型
type CalculateScore func(count Counts) uint64

// CalculateScoreSimple 代笔简易的组件评分计算函数
