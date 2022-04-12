package module

import (
	"net"
	"strconv"
	"testing"
)

var legalMIDs = []MID{}

func init() {
	for _, mt := range legalTypes {
		for mip := range legalIPMap {
			addr, _ := NewAddr("http", mip, 8080)
			mid, _ := GenMID(mt, DefaultSNGen.Get(), addr)
			legalMIDs = append(legalMIDs, mid)
		}
	}
}

var illegalMIDs = []MID{
	MID("D"),
	MID("DZ"),
	MID("D1|"),
	MID("D1|127.0.0.1:-1"),
	MID("D1|127.0.0.1:"),
	MID("D1|127.0.0.1"),
	MID("D1|127.0.0."),
	MID("D1|127"),
	MID("D1|127.0.0.0.1:8080"),
	MID("DZ|127.0.0.1:8080"),
	MID("A"),
	MID("AZ"),
	MID("A1|"),
	MID("A1|127.0.0.1:-1"),
	MID("A1|127.0.0.1:"),
	MID("A1|127.0.0.1"),
	MID("A1|127.0.0."),
	MID("A1|127"),
	MID("A1|127.0.0.0.1:8080"),
	MID("AZ|127.0.0.1:8080"),
	MID("P"),
	MID("PZ"),
	MID("P1|"),
	MID("P1|127.0.0.1:-1"),
	MID("P1|127.0.0.1:"),
	MID("P1|127.0.0.1"),
	MID("P1|127.0.0."),
	MID("P1|127"),
	MID("P1|127.0.0.0.1:8080"),
	MID("PZ|127.0.0.1:8080"),
	MID("M1|127.0.0.1:8080"),
}

func TestMIDGenAndSplit(t *testing.T) {
	addr, _ := NewAddr("http", "127.0.0.1", 8080)
	addrs := []net.Addr{nil, addr}
	for _, addr := range addrs {
		for _, mt := range legalTypes {
			expectedSN := DefaultSNGen.Get()
			mid, err := GenMID(mt, expectedSN, addr)
			if err != nil {
				t.Fatalf("生成模块ID时出错: %s (type: %s, sn: %d, addr: %s)", err, mt, expectedSN, addr)
			}

			expectedLetter := legalTypeLetterMap[mt]
			var expectedAddrStr string
			if addr != nil {
				expectedAddrStr = addr.String()
			}

			parts, err := SplitMID(mid)
			if err != nil {
				t.Fatalf("拆分 mid %q 时出错: %s", mid, err)
			}

			letter, snStr, addrStr := parts[0], parts[1], parts[2]
			if letter != expectedLetter {
				t.Fatalf("MID 中的类型字母不一致，预期: %s, 实际: %s", expectedLetter, letter)
			}

			sn, err := strconv.ParseUint(snStr, 10, 64)
			if err != nil {
				t.Fatalf("解析 SN: %s(snStr: %s)时出错", err, snStr)
			}

			if sn != expectedSN {
				t.Fatalf("MID 中的SN不一致，预期: %d, 实际: %d", expectedSN, sn)
			}

			if addrStr != expectedAddrStr {
				t.Fatalf("MID 中的地址字符串不一致。预期: %s, 实际: %s", expectedAddrStr, addrStr)
			}
		}
	}

	for _, addr := range addrs {
		for _, mt := range illegalTypes {
			mid, err := GenMID(mt, DefaultSNGen.Get(), addr)
			if err == nil {
				t.Fatalf("它仍然可以生成非法类型 %q!", mt)
			}

			if string(mid) != "" {
				t.Fatalf("它仍然可以生成具有非法类型 %q 的模块 ID %q!", mid, mt)
			}
		}
	}

	for _, illegalMID := range illegalMIDs {
		if _, err := SplitMID(illegalMID); err == nil {
			if _, err := SplitMID(illegalMID); err == nil {
				t.Fatalf("可拆分非法模块ID %q", illegalMID)
			}
		}
	}
}

func TestMIDLegal(t *testing.T) {
	var addr net.Addr
	var mid MID
	var err error

	for _, mt := range legalTypes {
		for mip := range legalIPMap {
			sn := DefaultSNGen.Get()
			addr, err = NewAddr("http", mip, 8080)
			if err == nil {
				mid, err = GenMID(mt, sn, addr)
			}

			if err != nil {
				t.Fatalf("判断MID : %s (type: %s, sn: %d, addr: %s) 的合法性时出错", err, mt, sn, addr)
			}

			if !LegalMID(mid) {
				t.Fatalf("生成的 MID %q 是合法的，但不会被检测到!", mid)
			}
		}
	}

	for _, illegalMID := range illegalMIDs {
		if LegalMID(illegalMID) {
			t.Fatalf("生成的MID %q 是非法的，但不会被检测到", illegalMID)
		}
	}
}
