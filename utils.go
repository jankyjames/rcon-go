package rcon

import (
	"errors"
	"os"
)

func isSyscallError(err error) (bool, *os.SyscallError) {
	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		return true, sysErr
	}

	return false, nil
}
