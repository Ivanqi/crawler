package buffer

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolNew(t *testing.T) {
	bufferCap := uint32(10)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)

	if err != nil {
		t.Fatalf("新建缓冲池时出错: %s"+"(bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}

	if pool == nil {
		t.Fatalf("无法创建缓冲池!")
	}

	if pool.BufferCap() != bufferCap {
		t.Fatalf("缓冲区容量不一致: 预期:%d, 当前:%d", bufferCap, pool.BufferCap())
	}

	if pool.MaxBufferNumber() != maxBufferNumber {
		t.Fatalf("最大缓冲区数不一致: 预期: %d, 实际: %d", maxBufferNumber, pool.MaxBufferNumber())
	}

	if pool.BufferNumber() != 1 {
		t.Fatalf("缓冲区数不一致: 预期: %d, 实际: %d", 1, pool.BufferNumber())
	}

	pool, err = NewPool(0, 1)
	if err == nil {
		t.Fatalf("新缓冲池为零时没有错误")
	}

	pool, err = NewPool(1, 0)
	if err == nil {
		t.Fatalf("新建一个最大缓冲区数为零的缓冲区时不会出错!")
	}
}

// addExtraDatum 用于在池已满时再放入一个数据
func addExtraDatum(pool Pool, datum interface{}) chan error {
	sign := make(chan error, 1)
	go func() {
		sign <- pool.Put(datum)
	}()
	return sign
}

func TestPoolPut(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)

	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("新建缓冲池时出错: %s"+"(bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}

	dataLen := bufferCap * maxBufferNumber
	data := make([]uint32, dataLen)
	for i := uint32(0); i < dataLen; i++ {
		data[i] = i
	}

	var count uint32
	var datum uint32
	for _, datum = range data {
		err := pool.Put(datum)
		if err != nil {
			t.Fatalf("将数据放入缓冲池时出错: %s, (数据: %d)", err, datum)
		}
		count++

		if pool.Total() != uint64(count) {
			t.Fatalf("数据总数不一致: 预期: %d, 实际: %d", count, pool.Total())
		}

		expectedBufferNumber := count / uint32(bufferCap)
		if count%uint32(bufferCap) != 0 {
			expectedBufferNumber++
		}

		if pool.BufferNumber() != expectedBufferNumber {
			t.Fatalf("缓冲区数不一致：预期: %d, 实际: %d (计数: %d)", expectedBufferNumber, pool.BufferNumber(), count)
		}
	}

	datum = dataLen
	select {
	case err := <-addExtraDatum(pool, datum):
		if err != nil {
			t.Fatalf("将日期放入缓冲池时出错: %s, (数据: %d)", err, datum)
		} else {
			t.Fatalf("它仍然可以为完整的缓冲池提供一个数据")
		}
	case <-time.After(time.Microsecond):
		t.Logf("超市！无法将数据放入完整的缓冲池")
	}
	pool.Close()
	err = pool.Put(datum)
	if err == nil {
		t.Fatalf("它仍然可以将数据放入已关闭的缓冲池! (数据: %d)", datum)
	}
}

func TestPoolPutInParallel(t *testing.T) {
	bufferCap := uint32(30)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("新建缓冲池时出错: %s"+"(bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}

	dataLen := bufferCap * maxBufferNumber
	data := make([]uint32, dataLen)
	for i := uint32(0); i < dataLen; i++ {
		data[i] = i
	}

	var count uint32
	testingFunc := func(datum interface{}, t *testing.T) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			err := pool.Put(datum)
			if err != nil {
				t.Fatalf("将数据放入缓冲池出错: %s (数据: %d)", err, datum)
			}
			atomic.AddUint32(&count, 1)
			currentCount := atomic.LoadUint32(&count)
			if uint64(currentCount) > pool.Total() {
				t.Fatalf("不一致的数据总数: %d > %d (old > new)", currentCount, pool.Total())
			}
		}
	}

	t.Run("Put in parallel(1)", func(t *testing.T) {
		for _, datum := range data[:dataLen/2] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum, t))
		}
	})

	t.Run("Put in parallel(2)", func(t *testing.T) {
		for _, datum := range data[dataLen/2:] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum, t))
		}
	})

	datum := dataLen
	select {
	case err := <-addExtraDatum(pool, datum):
		if err != nil {
			t.Fatalf("将数据放入缓冲池时出错: %s, (数据: %d)", err, datum)
		} else {
			t.Fatalf("它仍然可以为完整的缓冲池提供一个数据")
		}
	case <-time.After(time.Microsecond):
		t.Logf("超时! 无法将数据放入完整的缓冲池")
	}

	pool.Close()
}

// getExtraDatum 用于在池已空时再获取一个数据
func getExtraDatum(pool Pool) chan error {
	sign := make(chan error, 1)
	go func() {
		_, err := pool.Get() // 这条语句应该会一直阻塞
		sign <- err
	}()
	return sign
}

func TestPoolGet(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("新疆缓冲池出错: %s"+"(bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}

	dataLen := uint32(bufferCap * maxBufferNumber)
	for i := uint32(0); i < dataLen; i++ {
		pool.Put(i)
	}

	count := dataLen
	expectedBufferNumber := maxBufferNumber
	var datum uint32
	var ok bool
	for i := uint32(0); i < dataLen; i++ {
		d, err := pool.Get()
		if err != nil {
			t.Fatalf("从缓冲池获取数据时出错: %s", err)
		}

		datum, ok = d.(uint32)
		if !ok {
			t.Fatalf("给定类型不一致: 预期: %T, 实际: %T", datum, d)
		}

		if datum < 0 || datum >= dataLen {
			t.Fatalf("数据超出范围: 预期: [0, %d], 实际: %d", count, pool.Total())
		}
		count--
		if pool.Total() != uint64(count) {
			t.Fatalf("数据总数不一致: 预期: %d, 实际: %d", count, pool.Total())
		}

		if pool.BufferNumber() != expectedBufferNumber {
			t.Fatalf("缓冲区数不一致: 预期: %d, 实际: %d (count: %d))", expectedBufferNumber, pool.BufferNumber(), count)
		}
	}

	select {
	case err := <-getExtraDatum(pool):
		if err != nil {
			t.Fatalf("从缓冲池获取数据时出错: %s", err)
		} else {
			t.Fatalf("它仍然可以从空缓冲池中获取数据")
		}
	case <-time.After(time.Millisecond):
		t.Logf("超时，不能从空的缓冲池中获取数据")
	}
	datum = 0
	pool.Put(datum)
	pool.Close()
	_, err = pool.Get()
	if err == nil {
		t.Fatalf("它仍然可以从封闭的缓冲池中获取数据！")
	}
}

func TestPoolGetInParallel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("新建缓冲池出错: %s "+"(bufferCap: %d, maxBufferNumber: %d)", err, bufferCap, maxBufferNumber)
	}

	dataLen := uint32(bufferCap * maxBufferNumber)
	for i := uint32(0); i < dataLen; i++ {
		pool.Put(i)
	}

	count := dataLen
	testingFunc := func(t *testing.T) {
		t.Parallel()
		d, err := pool.Get()
		if err != nil {
			t.Fatalf("从缓冲池获取数据时出错: %s", err)
		}

		datum, ok := d.(uint32)
		if !ok {
			t.Fatalf("给定类型不一致: 预期: %T, 实际: %T", datum, d)
		}

		if datum < 0 || datum >= dataLen {
			t.Fatalf("数据超出范围: 预期: [0, %d), 实际: %d", dataLen, datum)
		}

		atomic.AddUint32(&count, ^uint32(0))
		currentCount := atomic.LoadUint32(&count)
		if uint64(currentCount) < pool.Total() {
			t.Fatalf("不一致的数据总数: %d < %d (old < new)", currentCount, pool.Total())
		}
	}

	t.Run("Get in parallel(1)", func(t *testing.T) {
		min := uint32(0)
		max := dataLen / 2
		for i := min; i < max; i++ {
			t.Run(fmt.Sprintf("Index=%d", i), testingFunc)
		}
	})
	t.Run("Get in parallel(2)", func(t *testing.T) {
		min := dataLen / 2
		max := dataLen
		for i := min; i < max; i++ {
			t.Run(fmt.Sprintf("Index=%d", i), testingFunc)
		}
	})
	select {
	case err := <-getExtraDatum(pool):
		if err != nil {
			t.Fatalf("从缓冲池获取数据时出错: %s", err)
		} else {
			t.Fatal("它仍然可以从空缓冲池中获取数据")
		}
	case <-time.After(time.Millisecond):
		t.Logf("超时，不能从空的缓冲池中获取数")
	}
	pool.Close()
}

func TestPoolPutAndGetInParallel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when new a buffer pool: %s "+
			"(bufferCap: %d, maxBufferNumber: %d)",
			err, bufferCap, maxBufferNumber)
	}
	dataLen := uint32(bufferCap * maxBufferNumber)
	maxPuttingNumber := dataLen + uint32(rand.Int63n(20))
	maxGettingNumber := dataLen + uint32(rand.Int63n(20))
	puttingCount := maxPuttingNumber
	gettingCount := maxGettingNumber
	marks := make([]uint32, maxPuttingNumber)
	var lock sync.Mutex
	t.Run("All in parallel", func(t *testing.T) {
		t.Run("Put1", func(t *testing.T) {
			t.Parallel()
			begin := uint32(0)
			end := maxPuttingNumber / 2
			for i := begin; i < end; i++ {
				if pool.Total() == uint64(dataLen) {
					datum := dataLen
					select {
					case err := <-addExtraDatum(pool, datum):
						if err != nil {
							t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)",
								err, datum)
						} else {
							t.Fatal("It still can put a datum to the full buffer pool!")
						}
					case <-time.After(time.Millisecond):
						t.Logf("Timeout! Couldn't put data to the full buffer pool.")
					}
					continue
				}
				err := pool.Put(i)
				if err != nil {
					t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)",
						err, i)
				}
				atomic.AddUint32(&puttingCount, ^uint32(0))
			}
		})
		t.Run("Put2", func(t *testing.T) {
			t.Parallel()
			begin := maxPuttingNumber / 2
			end := maxPuttingNumber
			for i := begin; i < end; i++ {
				if pool.Total() == uint64(dataLen) {
					datum := dataLen
					select {
					case err := <-addExtraDatum(pool, datum):
						if err != nil {
							t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)",
								err, datum)
						} else {
							t.Fatal("It still can put a datum to the full buffer pool!")
						}
					case <-time.After(time.Millisecond):
						t.Logf("Timeout! Couldn't put data to the full buffer pool.")
					}
					continue
				}
				err := pool.Put(i)
				if err != nil {
					t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)",
						err, i)
				}
				atomic.AddUint32(&puttingCount, ^uint32(0))
			}
		})
		t.Run("Get1", func(t *testing.T) {
			t.Parallel()
			max := dataLen/2 + 1
			for i := uint32(0); i < max; i++ {
				if pool.Total() == 0 {
					select {
					case err := <-getExtraDatum(pool):
						if err != nil {
							t.Fatalf("An error occurs when getting a datum from the buffer pool: %s",
								err)
							// } else {
							// 	t.Fatal("It still can get a datum from the empty buffer pool!")
						}
					case <-time.After(time.Millisecond):
						t.Logf("Timeout! Couldn't get data from the empty buffer pool.")
					}
					continue
				}
				d, err := pool.Get()
				if err != nil {
					t.Fatalf("An error occurs when getting a datum from the buffer pool: %s",
						err)
				}
				if d == nil &&
					atomic.LoadUint32(&puttingCount) == 0 &&
					pool.Total() != 0 {
					t.Fatalf("Get an empty datum! (total: %d)", pool.Total())
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
			max := dataLen/2 + 2
			for i := uint32(0); i < max; i++ {
				if pool.Total() == 0 {
					select {
					case err := <-getExtraDatum(pool):
						if err != nil {
							t.Fatalf("An error occurs when getting a datum from the buffer pool: %s",
								err)
							// } else {
							// 	t.Fatal("It still can get a datum from the empty buffer pool!")
						}
					case <-time.After(time.Millisecond):
						t.Logf("Timeout! Couldn't get data from the empty buffer pool.")
					}
					continue
				}
				d, err := pool.Get()
				if err != nil {
					t.Fatalf("An error occurs when getting a datum from the buffer pool: %s",
						err)
				}
				if d == nil &&
					atomic.LoadUint32(&puttingCount) == 0 &&
					pool.Total() != 0 {
					t.Fatalf("Get an empty datum! (total: %d)", pool.Total())
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
			t.Fatalf("Got the number more than once: %d", i)
		}
	}
	pool.Close()
}

func TestPoolCloseInParallel(t *testing.T) {
	bufferCap := uint32(20)
	maxBufferNumber := uint32(10)
	pool, err := NewPool(bufferCap, maxBufferNumber)
	if err != nil {
		t.Fatalf("An error occurs when new a buffer pool: %s "+
			"(bufferCap: %d, maxBufferNumber: %d)",
			err, bufferCap, maxBufferNumber)
	}
	dataLen := uint32(bufferCap * maxBufferNumber)
	maxNumber := dataLen / 2
	t.Run("Put", func(t *testing.T) {
		t.Parallel()
		for i := uint32(0); i < maxNumber; i++ {
			err := pool.Put(i)
			if err != nil && !pool.Closed() {
				t.Fatalf("An error occurs when putting a datum to the buffer pool: %s (datum: %d)",
					err, i)
			}
		}
	})
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		for i := uint32(0); i < maxNumber; i++ {
			_, err := pool.Get()
			if err != nil && !pool.Closed() {
				t.Fatalf("An error occurs when getting a datum from the buffer pool: %s (datum: %d)",
					err, i)
			}
		}
	})
	t.Run("Close", func(t *testing.T) {
		t.Parallel()
		time.Sleep(time.Millisecond)
		ok := pool.Close()
		if !ok {
			t.Fatal("Couldn't close the buffer pool!")
		}
		if !pool.Closed() {
			t.Fatalf("Inconsistent buffer pool status: expected closed: %v, actual closed: %v",
				true, pool.Closed())
		}
		ok = pool.Close()
		if ok {
			t.Fatal("It still can close the closed buffer pool!")
		}
		if !pool.Closed() {
			t.Fatalf("Inconsistent buffer pool status: expected closed: %v, actual closed: %v",
				true, pool.Closed())
		}
	})
}
