//go:build !windows

package rcon

import (
	"syscall"
)

// isRefused returns true when an error is returned do to the host not listening, which is typical during
// server startup.
func isRefused(err error) bool {
	is, sysErr := isSyscallError(err)
	return is && (sysErr.Err == syscall.ECONNREFUSED)
}

func isAborted(err error) bool {
	is, sysErr := isSyscallError(err)
	return is && (sysErr.Err == syscall.ECONNABORTED)
}
