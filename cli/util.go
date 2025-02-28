package cli

import (
	"os"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

func TerminalSupportsANSI() bool {
	// Falls es sich um ein echtes Terminal handelt
	if term.IsTerminal(int(os.Stdout.Fd())) {
		// Unter Windows muss ANSI explizit aktiviert werden
		if runtime.GOOS == "windows" {
			var mode uint32
			handle := windows.Handle(os.Stdout.Fd())
			if err := windows.GetConsoleMode(handle, &mode); err == nil {
				// ANSI aktivieren (falls nicht bereits aktiv)
				_ = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
				return true
			}
			return false
		}

		// Unter Unix können wir TERM überprüfen
		termEnv := os.Getenv("TERM")
		ansiTerms := []string{"xterm", "screen", "tmux", "rxvt", "vt100", "linux", "cygwin"}
		for _, term := range ansiTerms {
			if strings.Contains(termEnv, term) {
				return true
			}
		}
	}

	return false
}
