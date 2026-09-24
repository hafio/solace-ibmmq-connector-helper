//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// enableVirtualTerminal turns on ANSI escape processing for w, which Windows
// consoles need before the screen-clearing sequences --watch prints mean
// anything. Windows Terminal and PowerShell enable it for themselves, but a
// plain conhost window does not, and without it a redraw would print the escape
// bytes as text.
//
// Best-effort and silent: a handle that is not a console (a redirected file, a
// pipe into a pager) simply fails the call, which is the right outcome -- there
// is no screen to clear there either. Called on every redraw rather than once,
// because it is a single cheap syscall and that keeps the watch loop free of
// setup state.
//
// It uses syscall's lazy DLL loading rather than golang.org/x/sys/windows to
// avoid promoting that module to a direct dependency for two calls.
func enableVirtualTerminal(w *os.File) {
	const enableVirtualTerminalProcessing = 0x0004

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	handle := syscall.Handle(w.Fd())
	var mode uint32
	if ret, _, _ := getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode))); ret == 0 {
		return // not a console
	}
	if mode&enableVirtualTerminalProcessing != 0 {
		return // already on
	}
	setConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
}
