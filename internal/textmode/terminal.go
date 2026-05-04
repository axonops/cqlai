package textmode

import (
	readline "github.com/chzyer/readline"
)

// readOneRawViaReadline reads a single byte from the terminal using readline's
// own raw-mode machinery, so we never call term.MakeRaw while readline owns
// the terminal.
//
// Approach: between Readline() calls the terminal is in cooked mode.  We use
// rl.Config.FuncMakeRaw / rl.Config.FuncExitRaw — the same hooks readline
// itself uses — to enter raw mode, read one byte from rl.Config.Stdin (the
// FillableStdin wrapper readline already installed), then restore cooked mode.
// This avoids importing golang.org/x/term and avoids double raw-mode toggling.
//
// If rl is nil (e.g. in tests or batch mode) the function returns (0, false).
func readOneRawViaReadline(rl *readline.Instance) (byte, bool) {
	if rl == nil {
		return 0, false
	}

	cfg := rl.Config
	if err := cfg.FuncMakeRaw(); err != nil {
		// Could not enter raw mode; skip the prompt.
		return 0, false
	}
	defer func() { _ = cfg.FuncExitRaw() }()

	b := make([]byte, 1)
	_, err := cfg.Stdin.Read(b)
	if err != nil {
		return 0, false
	}
	return b[0], true
}
