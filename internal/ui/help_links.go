package ui

import (
	"os/exec"
	"regexp"
	"runtime"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Links in the help window.
//
// cqlai takes the mouse, so the terminal never sees a Ctrl+click and cannot
// open a link the way it would in a shell. cqlai opens it instead, with the
// desktop's own opener. Where there is none - over ssh, say - the link is
// copied, so it can be pasted into a browser.

// linkPattern is a link in a line of help.
var linkPattern = regexp.MustCompile(`https?://[^\s<>"']+`)

// linkOpenWait is how long to wait to hear whether the opener worked.
const linkOpenWait = 3 * time.Second

// linkNotOpenedMsg says the link could not be opened on this machine.
type linkNotOpenedMsg struct{ url string }

// linkAt is the link at a column of a line, if there is one there.
func linkAt(line string, col int) (string, bool) {
	text := ansi.Strip(line)
	for _, span := range linkPattern.FindAllStringIndex(text, -1) {
		url := trimLinkEnd(text[span[0]:span[1]])
		from := ansi.StringWidth(text[:span[0]])
		if col >= from && col < from+ansi.StringWidth(url) {
			return url, true
		}
	}
	return "", false
}

// trimLinkEnd drops the full stop or bracket a sentence puts after a link.
func trimLinkEnd(url string) string {
	for len(url) > 0 {
		switch url[len(url)-1] {
		case '.', ',', ';', ':', ')', ']', '!', '?':
			url = url[:len(url)-1]
		default:
			return url
		}
	}
	return url
}

// helpLinkAt is the link under a screen position in the help window.
func (m *MainModel) helpLinkAt(col, row int) (string, bool) {
	source, ok := m.helpSource()
	if !ok {
		return "", false
	}
	line, column, ok := m.docPosition(col, row)
	if !ok || line < 0 || line >= len(source.lines) {
		return "", false
	}
	return linkAt(source.lines[line], column)
}

// openLink opens a link in the desktop's browser, and copies it if that fails.
func (m *MainModel) openLink(url string) (*MainModel, tea.Cmd) {
	m.help.notice = "Opening " + url
	return m, func() tea.Msg {
		if openInBrowser(url) {
			return nil
		}
		return linkNotOpenedMsg{url: url}
	}
}

// linkNotOpened copies the link, and says so in the help window.
func (m *MainModel) linkNotOpened(msg linkNotOpenedMsg) (*MainModel, tea.Cmd) {
	m.lastCopied = msg.url
	if m.help.active {
		m.help.notice = "No browser here. The link is copied"
	}
	return m, tea.Batch(tea.SetClipboard(msg.url), writeSystemClipboard(msg.url))
}

// openInBrowser is the open itself, behind a variable so a test does not start
// a browser.
var openInBrowser = startBrowser

// startBrowser runs this desktop's opener, and reports whether it started.
func startBrowser(url string) bool {
	var command []string
	switch runtime.GOOS {
	case "darwin":
		command = []string{"open", url}
	case "windows":
		command = []string{"rundll32", "url.dll,FileProtocolHandler", url}
	default:
		command = []string{"xdg-open", url}
	}

	path, err := exec.LookPath(command[0])
	if err != nil {
		return false
	}
	// #nosec G204 - the opener is from the fixed table above, and the link is
	// one matched in cqlai's own help text.
	cmd := exec.Command(path, command[1:]...)
	if cmd.Start() != nil {
		return false
	}

	// Most openers hand the link over and exit. One that is still running
	// after a while has started a browser in the foreground, which is left to
	// run; one that exits with an error had nothing to open it with.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err == nil
	case <-time.After(linkOpenWait):
		return true
	}
}
