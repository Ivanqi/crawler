package module

import (
	"strconv"
	"testing"
)

var legalNetworkMap = map[string]struct{}{
	"http":  struct{}{},
	"https": struct{}{},
}

var legalIPMap = map[string]struct{}{
	"192.0.2.1":    struct{}{},
	"2001:db8::68": struct{}{},
}

func TestAddrNew(t *testing.T) {
	var network string
	var ip string
	port := uint64(8080)

	for network = range legalNetworkMap {
		for ip = range legalIPMap {
			addr, err := NewAddr(network, ip, port)
			if err != nil {
				t.Fatalf("创建地址时出错: %s (network: %s, ip: %s, port: %d)", err, network, ip, port)
			}

			if addr == nil {
				t.Fatal("不能创建地址!")
			}

			if addr.Network() != network {
				t.Fatalf("地址的网络不一致：预期:%s, 实际: %s", network, addr.Network())
			}

			expectedIPPort := ip + ":" + strconv.FormatUint(port, 10)
			if addr.String() != expectedIPPort {
				t.Fatalf("地址的IP/端口不一致： 预期: %s, 实际: %s", expectedIPPort, addr.String())
			}
		}
	}

	network = "tcp"
	_, err := NewAddr(network, ip, port)
	if err == nil {
		t.Fatalf("使用非法网络 %q! 创建缓冲区时没有错误!", network)
	}

	network = "http"
	ip = "127.0.0.0.1"
	_, err = NewAddr(network, ip, port)
	if err == nil {
		t.Fatalf("使用非法网络 %q! 创建缓冲区时没有错误！", network)
	}
}
