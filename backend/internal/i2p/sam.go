package i2p

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// DetectRouterMode attempts to detect whether local router is I2P or I2P+.
// It fetches the console root and looks for indicative strings. Returns "I2P", "I2P+", or error.
func DetectRouterMode(consoleURL string) (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(consoleURL + "/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		// heuristics:
		if strings.Contains(line, "I2P+") || strings.Contains(line, "I2P Plus") {
			return "I2P+", nil
		}
		if strings.Contains(line, "I2P Router Console") || strings.Contains(line, "I2P Router") {
			return "I2P", nil
		}
	}
	return "unknown", nil
}

// IsSAMAvailable probes assumed SAM address (host:port) and returns true if listening.
func IsSAMAvailable(samAddr string) bool {
	conn, err := net.DialTimeout("tcp", samAddr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// CreateSocksDialer returns a SOCKS5 dialer using golang.org/x/net/proxy.
// socksAddr example: "127.0.0.1:4444"
func CreateSocksDialer(socksAddr string) (proxy.Dialer, error) {
	return proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
}

// Quick check function that uses socks to connect to target
func DialThroughSocks(socksAddr, network, addr string, timeout time.Duration) (net.Conn, error) {
	dialer, err := CreateSocksDialer(socksAddr)
	if err != nil {
		return nil, err
	}
	type dialerIface interface {
		Dial(network, addr string) (net.Conn, error)
	}
	if d, ok := dialer.(dialerIface); ok {
		conn, err := d.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	// fallback to net.Dial (no socks)
	return net.DialTimeout(network, addr, timeout)
}

// Heuristic to see if local router exposes a JSON stats endpoint
func HasJSONStats(consoleURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(consoleURL + "/json")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
