//go:build windows

package discovery

import "syscall"

// On Windows SO_REUSEADDR lets another program steal the port: nothing is
// set, a busy port fails with a clear error instead.
func reuseAddr(_, _ string, _ syscall.RawConn) error { return nil }
