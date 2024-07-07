package rcon

import (
	"golang.org/x/sys/windows"
)

// isRefused returns true when an error is returned do to the host not listening, which is typical during
// server startup.
func isRefused(err error) bool {
	is, sysErr := isSyscallError(err)
	return is && (sysErr.Err == windows.WSAECONNREFUSED)
}

func isAborted(err error) bool {
	is, sysErr := isSyscallError(err)
	return is && (sysErr.Err == windows.WSAECONNABORTED)
}
