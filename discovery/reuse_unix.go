//go:build !windows

package discovery

import "syscall"

// reuseAddr lets the scan bind a discovery port that a vendor tool or a
// previous scan still holds in TIME_WAIT-like states, and share a multicast
// port with other listeners of the group.
func reuseAddr(_, _ string, c syscall.RawConn) error {
	var serr error
	err := c.Control(func(fd uintptr) {
		serr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})
	if err != nil {
		return err
	}
	return serr
}
