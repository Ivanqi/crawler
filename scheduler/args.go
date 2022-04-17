package scheduler

// Args 代表参数容器的接口类型
type Args interface {
	// Check 用于自检参数的有效性
	// 若结果值为nil，则说明未发现问题，否则就意味着自检未通过
	Check() error
}

// RequestArgs 代表请求相关的参数容器的类型
type RequestArgs interface {
	// AcceptedDomains 代表可以接受的URL的主域名的列表
	// URL主域名不在列表中的请求都会被忽略
	AcceptedDomains []string `json:"accepted_primary_domains"`
	// maxDepth 代表了需要被爬取的最大深度
	// 实际深度大于此值的请求都会被忽略
	MaxDepth uint32 `json:"max_depth"`
}

func (args *ReqRequestArgs)
