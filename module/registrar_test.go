package module

import (
	"fmt"
	"net"
	"testing"
)

func TestRegNew(t *testing.T) {
	registrar := NewRegistrar()
	if registrar == nil {
		t.Fatal("不能创建 registrar！")
	}
}

func TestRegRegister(t *testing.T) {
	mt := TYPE_DOWNLOADER
	ml := legalTypeLetterMap[mt]
	sn := DefaultSNGen.Get()
	addr, _ := NewAddr("http", "127.0.0.1", 8080)
	mid := MID(fmt.Sprintf(midTemplate, ml, sn, addr))

	// 测试无效组件实例的注册
	registrar := NewRegistrar()
	ok, err := registrar.Register(nil)
	if err == nil {
		t.Fatal("使用nil 模块注册实例时没有错误!")
	}

	if ok {
		t.Fatalf("仍然可以注册nil模块实例")
	}

	// 测试类型不匹配的组件实例的注册
	var m Module
	for t, f := range fakeModuleFuncMap {
		if t != mt {
			m = f(mid)
			break
		}
	}

	ok, err = registrar.Register(m)
	if err == nil {
		t.Fatalf("注册不匹配的模块实例没有错误! (type: %T)", m)
	}

	if ok {
		t.Fatalf("仍然可以注册不匹配的模块实例! (type: %T)", m)
	}

	var midsAll []MID
	for _, mt := range legalTypes {
		var midsByType []MID
		for mip := range legalIPMap {
			ml = legalTypeLetterMap[mt]
			sn = DefaultSNGen.Get()
			addr, _ = NewAddr("http", mip, 8080)
			mid = MID(fmt.Sprintf(midTemplate, ml, sn, addr))
			midsByType = append(midsByType, mid)
			midsAll = append(midsAll, mid)
			m = fakeModuleFuncMap[mt](mid)
			ok, err = registrar.Register(m)
			if err != nil {
				t.Fatalf("注册模块实例时出错: %s (MID: %s)", err, mid)
			}

			if !ok {
				t.Fatalf("无法使用MID %q 来注册模块实例!", mid)
			}

			// 测试重复MID的注册
			ok, err = registrar.Register(m)
			if err != nil {
				t.Fatalf("注册模块实例时出错: %s (MID: %s)", err, mid)
			}

			if ok {
				t.Fatalf("仍然可以重复注册具有相同MID %q 的模块实例!", mid)
			}
		}

		modules, err := registrar.GetAllByType(mt)
		if err != nil {
			t.Fatalf("获取所有模块实例时出错: %s (type : %s)", err, mt)
		}

		for _, mid := range midsByType {
			if _, ok := modules[mid]; !ok {
				t.Fatalf("未找到模块实例! (MID: %s, type: %s)", mid, mt)
			}
		}
	}

	modules := registrar.GetAll()
	for _, mid := range midsAll {
		if _, ok := modules[mid]; !ok {
			t.Fatalf("未找到模块实例! (MID: %s)", mid)
		}
	}

	for _, mt := range illegalTypes {
		sn := DefaultSNGen.Get()
		addr, _ := NewAddr("http", "127.0.0.1", 8080)
		ml := legalTypeLetterMap[mt]
		mid := MID(fmt.Sprintf(midTemplate, ml, sn, addr))
		m := NewFakeDownloader(mid, nil)
		ok, err := registrar.Register(m)
		if err == nil {
			t.Fatalf("使用非法类型 %q! 注册模块实例时没有错误!", mt)
		}

		if ok {
			t.Fatalf("仍然可以注册具有非法类型 %q 的模块实例！", mt)
		}
	}
}

func TestModuleUnregister(t *testing.T) {
	registrar := NewRegistrar()
	var mids []MID
	for _, mt := range legalTypes {
		for mip := range legalIPMap {
			sn := DefaultSNGen.Get()
			addr, _ := NewAddr("http", mip, 8080)
			mid, err := GenMID(mt, sn, addr)
			if err != nil {
				t.Fatalf("生成模块ID时出错: %s (type: %s, sn: %d, addr: %s)", err, mt, sn, addr)
			}

			m := fakeModuleFuncMap[mt](mid)
			_, err = registrar.Register(m)
			if err != nil {
				t.Fatalf("注册模块实例时出错: %s(type: %s, sn: %d, 地址: %s)", err, mt, sn, addr)
			}

			mids = append(mids, mid)
		}
	}

	for _, mid := range mids {
		ok, err := registrar.Unregister(mid)
		if err != nil {
			t.Fatalf("注销模块实例时出错: %s (mid: %s)", err, mid)
		}

		if !ok {
			t.Fatalf("无法注销模块实例! (MID: %s)", mid)
		}
	}

	// 注销未注册的组件
	for _, mid := range mids {
		ok, err := registrar.Unregister(mid)
		if err != nil {
			t.Fatalf("注销模块实例时出错: %s (MID: %s)", err, mid)
		}

		if ok {
			t.Fatalf("它仍然可以注销不存在的模块实例! (MID: %s)", mid)
		}
	}

	for _, illegalMID := range illegalMIDs {
		ok, err := registrar.Unregister(illegalMID)
		if err == nil {
			t.Fatalf("注销具有非法 MID %q 的模块实例时没有错误!", illegalMID)
		}

		if ok {
			t.Fatalf("仍然可以注销具有非法MID %q 的模块实例", illegalMID)
		}
	}
}

func TestModuleGet(t *testing.T) {
	registrar := NewRegistrar()
	mt := illegalTypes[0]
	m1, err := registrar.Get(mt)
	if err == nil {
		t.Fatalf("获取具有非法类型 %q 的模块实例时没有错误!", mt)
	}

	if m1 != nil {
		t.Fatalf("仍然可以获得非法类型 %q 的模块实例！", mt)
	}

	mt = TYPE_DOWNLOADER
	m1, err = registrar.Get(mt)
	if err == nil {
		t.Fatal("获取不存在的模块实例时没有错误!")
	}

	if m1 != nil {
		t.Fatalf("仍然可以获得不存在的模块实例!")
	}

	addr, _ := NewAddr("http", "127.0.0.1", 8080)
	mid := MID(fmt.Sprintf(midTemplate, legalTypeLetterMap[mt], DefaultSNGen.Get(), addr))
	m := defaultFakeModuleMap[mt]
	_, err = registrar.Register(m)
	if err != nil {
		t.Fatalf("注册模块实例时出错: %s (mid: %s)", err, mid)
	}

	m1, err = registrar.Get(mt)
	if err != nil {
		t.Fatalf("获取模块实例时出错: %s (mid: %s)", err, mid)
	}

	if m1 == nil {
		t.Fatalf("无法获取 MID %q 的模块实例", mid)
	}

	if m1.ID() != m.ID() {
		t.Fatalf("不一致的MID: 预期: %s, 实例: %s", m.ID(), m1.ID())
	}
}

func TestModuleAllInParallel(t *testing.T) {
	baseSize := 1000
	basePort := 8000
	legalTypesLen := len(legalTypes)
	sLen := baseSize * legalTypesLen
	types := make([]Type, sLen)
	sns := make([]uint64, sLen)
	addrs := make([]net.Addr, sLen)
	for i := 0; i < sLen; i++ {
		types[i] = legalTypes[i%legalTypesLen]
		port := uint64(basePort + basePort%legalTypesLen)
		addrs[i], _ = NewAddr("http", "127.0.0.1", port)
		sns[i] = DefaultSNGen.Get()
	}
	registrar := NewRegistrar()
	t.Run("All in parallel", func(t *testing.T) {
		t.Run("Register", func(t *testing.T) {
			t.Parallel()
			for i, addr := range addrs {
				mt := types[i]
				sn := DefaultSNGen.Get()
				mid, err := GenMID(mt, sn, addr)
				if err != nil {
					t.Fatalf("生成模块ID时出错: %s (type: %s, sn: %d, addr: %s)", err, mt, sn, addr)
				}

				m := fakeModuleFuncMap[mt](mid)
				_, err = registrar.Register(m)
				if err != nil {
					t.Fatalf("注册模块实例时出错: %s (type: %s, sn: %d, addr: %s)", err, mt, sn, addr)
				}
			}
		})
		t.Run("Unregister", func(t *testing.T) {
			t.Parallel()
			for i, addr := range addrs {
				mt := types[i]
				sn := sns[i]
				mid, _ := GenMID(mt, sn, addr)
				_, err := registrar.Unregister(mid)
				if err != nil {
					t.Fatalf("注销模块实例时候出错: %s (mid: %s)", err, mid)
				}
			}
		})
		t.Run("Get", func(t *testing.T) {
			t.Parallel()
			for _, mt := range types {
				m, err := registrar.Get(mt)
				if err != nil && err != ErrNotFoundModuleInstance {
					t.Fatalf("获取模块实例时出错: %s (mid: %s)", err, m.ID())
				}
			}
		})
	})
}
