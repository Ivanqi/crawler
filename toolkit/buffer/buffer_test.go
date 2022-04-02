package buffer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBufferNew(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)

	if err != nil {
		t.Fatalf("新建缓冲器时出错：%s (大小：%d)", err, size)
	}

	if buf == nil {
		t.Fatalf("不能创建 Buffer")
	}

	if buf.Cap() != size {
		t.Fatalf("缓冲器上限不一致，预期为: %d, 实际为 %d", size, buf.Cap())
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
		t.Fatalf("新建缓冲器出错: %s (大小: %d)", err, size)
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
			t.Fatalf("将数据放入缓冲器时出错: %s, (datum: %d)", err, datum)
		}

		if !ok {
			t.Fatalf("datum 无法存储在buffer中! (datum: %d)", datum)
		}
		count++
		if buf.Len() != count {
			t.Fatalf("缓冲器Len 不一致. 预期: %d, 实际: %d", count, buf.Len())
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
		t.Fatalf("它仍然可以将datum放入关闭缓冲器! (datum: %d)", datum)
	}
}

func TestBufferPutInParallel(t *testing.T) {
	size := uint32(22)
	bufferSize := uint32(20)
	buf, err := NewBuffer(bufferSize)
	if err != nil {
		t.Fatalf("新建缓冲器时出错：%s (大小：%d)", err, size)
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
				t.Fatalf("将datum放入缓冲器时出错: %s (datum: %d)", err, datum)
			}

			if !ok && buf.Len() < buf.Cap() {
				t.Fatalf("不能把 datum 存入 缓冲器! (datum: %d)", datum)
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
		t.Fatalf("缓冲器Len 不一致，预期: %d, 实际: %d", buf.Cap(), buf.Len())
	}
}

func TestBufferGet(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)
	if err != nil {
		t.Fatalf("新建缓冲器时出错：%s (大小：%d)", err, size)
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
			t.Fatalf("从缓冲器获取datum出错: %s", err)
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
			t.Fatalf("缓冲器Len 不一致. 预期: %d, 实际: %d", count, buf.Len())
		}
	}

	d, err := buf.Get()
	if err != nil {
		t.Fatalf("从缓冲器中获取数据失败: %s", err)
	}

	if d != nil {
		t.Fatalf("它仍然可以从空缓冲器中获取数据!")
	}

	datum = 0
	buf.Put(datum)
	buf.Close()

	_, err = buf.Get()
	if err != nil {
		t.Fatalf("从缓冲器获取日期时出错: %s", err)
	}

	_, err = buf.Get()
	if err == nil {
		t.Fatal("它仍然可以从封闭的缓冲器中获取数据!")
	}
}

func TestBufferPutAndGetInParallel(t *testing.T) {
	bufferSize := uint32(50)
	buf, err := NewBuffer(bufferSize)
	if err != nil {
		t.Fatalf("新建缓冲器时出错：%s (大小：%d)", err, bufferSize)
	}

	maxPuttingNumber := bufferSize + uint32(rand.Int31n(20))
	maxGettingNumber := bufferSize + uint32(rand.Int31n(20))

	puttingCount := maxPuttingNumber
	gettingCount := maxGettingNumber

	marks := make([]uint8, maxPuttingNumber)
	var lock sync.Mutex

	t.Run("All in parallel", func(t *testing.T) {
		t.Run("Put1", func(t *testing.T) {
			t.Parallel()
			begin := uint32(0)
			end := maxPuttingNumber / 2
			for i := begin; i < end; i++ {
				ok, err := buf.Put(i)
				if err != nil {
					t.Fatalf("将数据放入缓冲器出错: %s (数据: %d)", err, i)
				}

				if !ok && atomic.LoadUint32(&gettingCount) == 0 && buf.Len() < buf.Cap() {
					t.Fatalf("无法将数据放入缓冲器! (数据: %d)", i)
				}

				atomic.AddUint32(&puttingCount, ^uint32(0))
			}
		})

		t.Run("Put2", func(t *testing.T) {
			t.Parallel()
			begin := maxGettingNumber / 2
			end := maxPuttingNumber
			for i := begin; i < end; i++ {
				ok, err := buf.Put(i)
				if err != nil {
					t.Fatalf("将数据放入缓冲器出错: %s, (数据: %d)", err, i)
				}

				if !ok && atomic.LoadUint32(&gettingCount) == 0 && buf.Len() < buf.Cap() {
					t.Fatalf("无法将数据放入缓冲器! (数据: %d)", i)
				}

				atomic.AddUint32(&puttingCount, ^uint32(0))
			}
		})

		t.Run("Get1", func(t *testing.T) {
			t.Parallel()
			max := bufferSize/2 + 1
			for i := uint32(0); i < max; i++ {
				d, err := buf.Get()
				if err != nil {
					t.Fatalf("从缓冲器获取数据出错: %s", err)
				}

				if d == nil && atomic.LoadUint32(&puttingCount) == 0 && buf.Len() != 0 {
					t.Fatalf("得到一个空的数据! (len: %d)", buf.Len())
				}

				atomic.AddUint32(&gettingCount, ^uint32(0))
				if d != nil {
					datum := d.(uint32)
					lock.Lock()
					marks[int(datum)]++
					lock.Unlock()
				}
			}
		})

		t.Run("Get2", func(t *testing.T) {
			t.Parallel()
			max := bufferSize/2 + 2
			for i := uint32(0); i < max; i++ {
				d, err := buf.Get()
				if err != nil {
					t.Fatalf("从缓冲器获取数据出错: %s", err)
				}

				if d == nil && atomic.LoadUint32(&puttingCount) == 0 && buf.Len() != 0 {
					t.Fatalf("获取空的数据! (len: %d)", buf.Len())
				}

				atomic.AddUint32(&gettingCount, ^uint32(0))
				if d != nil {
					datum := d.(uint32)
					lock.Lock()
					marks[int(datum)]++
					lock.Unlock()
				}
			}
		})
	})

	for i, m := range marks {
		if m > 1 {
			t.Fatalf("多次获取数据: %d", i)
		}
	}
}

func TestBufferCloseInParallel(t *testing.T) {
	bufferSize := uint32(100)
	buf, err := NewBuffer(bufferSize)
	if err != nil {
		t.Fatalf("新建缓冲器时出错：%s (大小：%d)", err, bufferSize)
	}
	maxNumber := bufferSize + uint32(rand.Int31n(100))
	t.Run("Put", func(t *testing.T) {
		t.Parallel()
		for i := uint32(0); i < maxNumber; i++ {
			_, err := buf.Put(i)
			if err != nil && !buf.Closed() {
				t.Fatalf("将数据放入缓冲器出错: %s, (数据: %d)", err, i)
			}
			if err == nil && buf.Closed() {
				t.Fatalf("仍然可以将数据放入封闭缓冲器! (数据: %d)", i)
			}
		}
	})
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		max := bufferSize/2 + 1
		for i := uint32(0); i < max; i++ {
			_, err := buf.Get()
			if err != nil && !buf.Closed() {
				t.Fatalf("从缓冲器获取数据时错误: %s, (数据: %d)", err, i)
			}
			if buf.Closed() {
				if _, err = buf.Get(); err == nil {
					t.Fatalf("仍然可以将数据放入封闭缓冲器! (数据: %d)", i)
				}
			}
		}
	})
	t.Run("Close", func(t *testing.T) {
		t.Parallel()
		time.Sleep(time.Millisecond)
		ok := buf.Close()
		if !ok {
			t.Fatal("不能关闭缓冲器")
		}

		if !buf.Closed() {
			t.Fatalf("缓冲器状态不一致: 预期关闭: %v, 实际关闭: %v", true, buf.Closed())
		}

		ok = buf.Close()
		if ok {
			t.Fatal("仍然不能关闭缓冲器！")
		}
		if !buf.Closed() {
			t.Fatalf("缓冲器状态不一致: 预期关闭: %v, 实际关闭: %v", true, buf.Closed())
		}
	})
}
