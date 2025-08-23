package server

import (
    "bytes"
    "encoding/binary"
    "net"
    "time"
)

// ProbeSRCDSInfo sends an A2S_INFO query to the given host:port and returns true if response received.
// https://developer.valvesoftware.com/wiki/Server_queries#A2S_INFO
func ProbeSRCDSInfo(addr string, timeout time.Duration) bool {
    conn, err := net.DialTimeout("udp", addr, timeout)
    if err != nil { return false }
    defer conn.Close()

    // A2S_INFO: 0xFF 0xFF 0xFF 0xFF 'T' "Source Engine Query\0"
    req := append([]byte{0xFF, 0xFF, 0xFF, 0xFF, 'T'}, []byte("Source Engine Query\x00")...)
    _ = conn.SetDeadline(time.Now().Add(timeout))
    if _, err := conn.Write(req); err != nil { return false }

    buf := make([]byte, 1400)
    n, err := conn.Read(buf)
    if err != nil || n < 6 { return false }
    // Basic check: header 0xFF 0xFF 0xFF 0xFF 0x49
    if n >= 5 && bytes.Equal(buf[:4], []byte{0xFF, 0xFF, 0xFF, 0xFF}) && buf[4] == 0x49 {
        // Optionally parse some fields (protocol, name length, etc.) – not required for health boolean
        _ = binary.LittleEndian
        return true
    }
    return false
}
