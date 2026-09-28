//go:build windows

package output

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode  = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode  = kernel32.NewProc("SetConsoleMode")
	procSetConsoleOutCP = kernel32.NewProc("SetConsoleOutputCP")
	enableVTProcessing  = uint32(0x0004)
	utf8CodePage        = uintptr(65001)
)

// enableVirtualTerminal turns on ANSI escape handling for a console.
func enableVirtualTerminal(f *os.File) bool {
	var mode uint32
	h := f.Fd()
	if r, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	if mode&enableVTProcessing != 0 {
		return true
	}
	r, _, _ := procSetConsoleMode.Call(h, uintptr(mode|enableVTProcessing))
	return r != 0
}

// PrepareConsole switches the console to UTF-8 so symbols such as ✓ render.
func PrepareConsole() {
	if procSetConsoleOutCP.Find() == nil {
		procSetConsoleOutCP.Call(utf8CodePage)
	}
}
