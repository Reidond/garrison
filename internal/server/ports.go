package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// ParsePorts parses YAML ports like "27015/udp" or "27015" and returns tuples.
func ParsePorts(specs []string) ([]struct {
	Port  int
	Proto string
}, error) {
	out := []struct {
		Port  int
		Proto string
	}{}
	for _, s := range specs {
		proto := "tcp"
		if strings.Contains(s, "/") {
			parts := strings.SplitN(s, "/", 2)
			s = parts[0]
			proto = strings.ToLower(parts[1])
		}
		p, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %s", s)
		}
		out = append(out, struct {
			Port  int
			Proto string
		}{p, proto})
	}
	return out, nil
}

// ProbePort tries to connect to the given TCP port on localhost with timeout.
func ProbePortTCP(port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return false
}
