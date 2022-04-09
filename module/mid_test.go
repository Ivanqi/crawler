package module

import (
	"net"
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
				t.Fatalf("生成模块ID时出错: %s (类型: %s, sn: %d)")
			}
		}
	}
}
