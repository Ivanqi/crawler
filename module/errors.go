package module

import "errors"

// ErrNotFoundModuleInstance 代表未找到组件实例的错误类型
var ErrNotFoundModuleInstance = errors.New("未找到模块实例")
