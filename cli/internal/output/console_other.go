//go:build !windows

package output

import "os"

func enableVirtualTerminal(*os.File) bool { return true }

// PrepareConsole is a no-op outside Windows.
func PrepareConsole() {}
