package ui

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Reading the clipboard of the machine cqlai is running on.
//
// A right click asks the terminal first, through OSC 52, which is the only
// route that works through ssh and tmux. Most terminals refuse it - a program
// that can read the clipboard can read what you copied out of a password
// manager - and then there is nothing to paste but what cqlai itself copied,
// which is no use for text copied from another application.
//
// So it also asks the machine, the way every other program does: the same
// command line the desktop's own clipboard tool offers. On a terminal running
// against a local desktop - including WSL, where the clipboard belongs to
// Windows - that answers. Over ssh with no forwarding it does not, and nothing
// else could.

// clipboardWait bounds the command. A clipboard read that has not answered in
// this long is not going to.
const clipboardWait = 2 * time.Second

// systemClipboardMsg is what the machine says its clipboard holds.
type systemClipboardMsg struct {
	request int
	text    string
}

// clipboardReaders are the commands that can say what the clipboard holds, in
// the order they are tried.
//
// Wayland first, then X, then macOS, then Windows - which is also how WSL
// reaches the clipboard it shares with the desktop it runs on.
func clipboardReaders() [][]string {
	powershell := []string{"powershell.exe", "-NoProfile", "-Command", "Get-Clipboard"}

	switch runtime.GOOS {
	case "darwin":
		return [][]string{{"pbpaste"}}
	case "windows":
		return [][]string{powershell}
	}

	return [][]string{
		{"wl-paste", "--no-newline"},
		{"xclip", "-selection", "clipboard", "-out"},
		{"xsel", "--clipboard", "--output"},
		powershell,
	}
}

// readSystemClipboard asks the machine what its clipboard holds.
func readSystemClipboard(request int) tea.Cmd {
	return func() tea.Msg {
		return systemClipboardMsg{request: request, text: systemClipboard()}
	}
}

// systemClipboard is the clipboard's contents, or "" when nothing here can say.
//
// Failures are silent on purpose: this runs because someone right-clicked, and
// the answer to "your terminal will not say and this machine has no clipboard
// tool" is to paste what cqlai copied, not to write an error into the console
// every time. What is read is never logged - it is the clipboard.
func systemClipboard() string {
	for _, reader := range clipboardReaders() {
		path, err := exec.LookPath(reader[0])
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), clipboardWait)
		// #nosec G204 - the command and its arguments are from the fixed table
		// above; only which of them exists on this machine varies.
		out, err := exec.CommandContext(ctx, path, reader[1:]...).Output()
		cancel()
		if err != nil {
			continue
		}

		// Windows hands it back with CRLF line endings and a trailing newline
		// the clipboard itself does not have.
		if text := strings.TrimRight(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n"); text != "" {
			return text
		}
	}
	return ""
}
