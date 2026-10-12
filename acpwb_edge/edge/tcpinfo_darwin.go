package edge

import (
	"net"
	"time"

	"golang.org/x/sys/unix"
)

// kernelRTT returns the kernel's smoothed round-trip time for a TCP socket
// (millisecond resolution on macOS).
func kernelRTT(c net.Conn) time.Duration {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return 0
	}
	sc, err := tc.SyscallConn()
	if err != nil {
		return 0
	}
	var rtt time.Duration
	sc.Control(func(fd uintptr) {
		if info, err := unix.GetsockoptTCPConnectionInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_CONNECTION_INFO); err == nil {
			rtt = time.Duration(info.Srtt) * time.Millisecond
		}
	})
	return rtt
}
