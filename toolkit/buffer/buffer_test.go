package buffer

import (
	"fmt"
	"testing"
)

func TestBufferNew(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)

	if err != nil {
		t.Fatalf("新建缓冲区时出错：%s (大小：%d)", err, size)
	}

	if buf == nil {
		t.Fatalf("不能创建 Buffer")
	}

	if buf.Cap() != size {
		t.Fatalf("缓冲区上限不一致，预期为: %d, 实际为 %d", size, buf.Cap())
	}

	buf, err = NewBuffer(0)
	if err == nil {
		t.Fatalf("当新的缓冲区大小为零时没有错误!")
	}
}

func TestBufferPut(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)
	if err != nil {
		t.Fatalf("新建缓冲区出错: %s (大小: %d)", err, size)
	}

	data := make([]uint32, size)
	for i := uint32(0); i < size; i++ {
		data[i] = i
	}

	var count uint32
	var datum uint32
	for _, datum = range data {
		ok, err := buf.Put(datum)
		if err != nil {
			t.Fatalf("将数据放入缓冲区时出错: %s, (datum: %d)", err, datum)
		}

		if !ok {
			t.Fatalf("datum 无法存储在buffer中! (datum: %d)", datum)
		}
		count++
		if buf.Len() != count {
			t.Fatalf("缓冲区Len 不一致. 预期: %d, 实际: %d", count, buf.Len())
		}
	}

	datum = size
	ok, err := buf.Put(datum)
	if err != nil {
		t.Fatalf("将数据放入缓冲区时出错: %s, (datum: %d)", err, datum)
	}

	if ok {
		t.Fatalf("它仍然可以将datum放入完整的缓冲区! (datum: %d)", datum)
	}

	buf.Close()
	_, err = buf.Put(datum)
	if err == nil {
		t.Fatalf("它仍然可以将datum放入关闭缓冲区! (datum: %d)", datum)
	}
}

func TestBufferPutInParallel(t *testing.T) {
	size := uint32(22)
	bufferSize := uint32(20)
	buf, err := NewBuffer(bufferSize)
	if err != nil {
		t.Fatalf("新建缓冲区时出错：%s (大小：%d)", err, size)
	}

	data := make([]uint32, size)
	for i := uint32(0); i < size; i++ {
		data[i] = i
	}

	testingFunc := func(datum interface{}, t *testing.T) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			ok, err := buf.Put(datum)
			if err != nil {
				t.Fatalf("将datum放入缓冲区时出错: %s (datum: %d)", err, datum)
			}

			if !ok && buf.Len() < buf.Cap() {
				t.Fatalf("不能把 datum 存入 缓冲区! (datum: %d)", datum)
			}
		}
	}

	t.Run("Put in parallel(1)", func(t *testing.T) {
		for _, datum := range data[:size/2] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum, t))
		}
	})

	t.Run("Put in parallel(2)", func(t *testing.T) {
		for _, datum := range data[size/2:] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum, t))
		}
	})

	if buf.Len() != buf.Cap() {
		t.Fatalf("缓冲区Len 不一致，预期: %d, 实际: %d", buf.Cap(), buf.Len())
	}
}

func TestBufferGet(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)
	if err != nil {
		t.Fatalf("新建缓冲区时出错：%s (大小：%d)", err, size)
	}

	for i := uint32(0); i < size; i++ {
		buf.Put(i)
	}

	count := size
	var datum uint32
	var ok bool
	for i := uint32(0); i < size; i++ {
		d, err := buf.Get()
		if err != nil {
			t.Fatalf("从缓冲区获取datum出错: %s", err)
		}
		datum, ok = d.(uint32)
		if !ok {
			t.Fatalf("给定类型不一致。预期: %T, 实际: %T", datum, d)
		}

		if datum != i {
			t.Fatalf("不一致的数据. 预期: %#v, 实际: %#v", i, datum)
		}
		count--

		if buf.Len() != count {
			t.Fatalf("缓冲区Len 不一致. 预期: %d, 实际: %d", count, buf.Len())
		}
	}

	d, err := buf.Get()
	if err != nil {
		t.Fatalf("从缓冲区中获取数据失败: %s", err)
	}

	if d != nil {
		t.Fatalf("它仍然可以从空缓冲区中获取数据!")
	}

	datum = 0
	buf.Put(datum)
	buf.Close()

	_, err = buf.Get()
	if err != nil {
		t.Fatalf("从缓冲区获取日期时出错: %s", err)
	}

	_, err = buf.Get()
	if err == nil {
		t.Fatal("它仍然可以从封闭的缓冲区中获取数据!")
	}
}
