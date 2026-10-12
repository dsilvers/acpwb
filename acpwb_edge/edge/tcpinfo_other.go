//go:build !linux && !darwin

package edge

import (
	"net"
	"time"
)

func kernelRTT(net.Conn) time.Duration { return 0 }
