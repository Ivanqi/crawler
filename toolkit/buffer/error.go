package buffer

import "error"

// ErrClosedBufferPool 表示缓冲池已经关闭的错误的变量
var ErrClosedBufferPool = errors.New("closed buffer pool")

// ErrClosedBuffer 表示缓冲器已关闭的错误的变量
var ErrClosedBuffer = errors.New("closed buffer")
