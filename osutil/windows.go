package osutil

import (
		"syscall"
	"golang.org/x/sys/windows"
)


func SupportsANSI() bool {
	// Get handle to standard output (console)
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return false
	}

	// Get current console mode
	var mode uint32
	err = windows.GetConsoleMode(windows.Handle(handle), &mode)
	if err != nil {
		return false
	}

	// Check if ENABLE_VIRTUAL_TERMINAL_PROCESSING is set
	return mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0
}