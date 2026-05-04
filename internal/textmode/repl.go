package textmode

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	readline "github.com/chzyer/readline"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
	"github.com/axonops/cqlai/internal/ui/completion"
)

// Run starts the text-mode REPL. It blocks until the user exits.
//
// The session and sessionMgr must already be connected. opts configures
// output format, history, etc.
func Run(
	_ context.Context,
	sess *db.Session,
	sessMgr *session.Manager,
	opts Options,
) error {
	// Resolve history file path.
	histFile := opts.HistoryFile
	if histFile == "" {
		histFile = os.Getenv("CQLAI_HISTORY_FILE")
	}
	if histFile == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			histFile = filepath.Join(home, ".cqlai_history")
		}
	}

	// Build the completion engine and readline completer.
	engine := completion.NewCompletionEngine(sess, sessMgr)
	completer := NewReadlineCompleter(engine)

	// Build the readline config.
	rlCfg := &readline.Config{
		Prompt:            buildPrompt(sessMgr.CurrentKeyspace()),
		HistoryFile:       histFile,
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	}

	rl, err := readline.NewEx(rlCfg)
	if err != nil {
		return fmt.Errorf("readline init: %w", err)
	}
	defer func() { _ = rl.Close() }()

	// Print the startup banner.
	PrintBanner(os.Stdout, sess, opts.Version)

	// Initialise the router with the session manager (idempotent if already done).
	router.InitRouter(sessMgr)

	p := newPrinter(opts)
	p.rl = rl
	buf := &InputBuffer{}

	for {
		// Update the prompt to reflect the current keyspace.
		rl.SetPrompt(buildPrompt(sessMgr.CurrentKeyspace()))

		// Update completer context with any accumulated buffer lines.
		if buf.IsEmpty() {
			completer.SetBufferPrefix("")
		} else {
			completer.SetBufferPrefix(buf.Text())
		}

		line, readErr := rl.Readline()

		if readErr != nil {
			if errors.Is(readErr, readline.ErrInterrupt) {
				// Ctrl-C: clear the buffer, do not exit.
				if !buf.IsEmpty() {
					fmt.Fprintln(os.Stderr, "(cancelled)")
				}
				buf.Reset()
				continue
			}
			if errors.Is(readErr, io.EOF) {
				// Ctrl-D on empty input: exit.
				fmt.Fprintln(os.Stdout, "exit")
				break
			}
			return readErr
		}

		// Trim trailing whitespace (but preserve inner content).
		trimmed := strings.TrimSpace(line)

		// Empty line: if buffer is empty, do nothing; otherwise continue accumulating
		// (empty line alone does not dispatch).
		if trimmed == "" {
			if buf.IsEmpty() {
				continue
			}
			// Don't add a blank continuation line — just keep waiting.
			rl.SetPrompt(continuationPrompt(sessMgr.CurrentKeyspace()))
			continue
		}

		buf.Add(trimmed)

		// Meta-commands are dispatched immediately (no semicolon required).
		if !buf.IsEmpty() && IsMeta(buf.lines[0]) && len(buf.lines) == 1 {
			stmt := buf.Text()
			buf.Reset()
			exit, result := dispatch(p, sess, sessMgr, stmt)
			if result != nil {
				fmt.Fprintln(os.Stderr, result.Error())
			}
			if exit {
				return nil
			}
			continue
		}

		// Check whether the buffer now contains a complete statement.
		if buf.IsComplete() {
			stmt := buf.Text()
			buf.Reset()
			exit, result := dispatch(p, sess, sessMgr, stmt)
			if result != nil {
				fmt.Fprintln(os.Stderr, result.Error())
			}
			if exit {
				return nil
			}
		} else {
			// Incomplete: switch to continuation prompt.
			rl.SetPrompt(continuationPrompt(sessMgr.CurrentKeyspace()))
		}
	}

	return nil
}

// parseUseKeyspace returns the keyspace name if stmt is a USE statement,
// otherwise returns "".  Handles both unquoted (USE myks;) and double-quoted
// (USE "MyKeyspace";) forms, matching the behaviour of the db.executor layer.
func parseUseKeyspace(stmt string) string {
	trimmed := strings.TrimSpace(stmt)
	upper := strings.ToUpper(trimmed)
	if !strings.HasPrefix(upper, "USE ") {
		return ""
	}
	// Extract everything after "USE "
	rest := strings.TrimSpace(trimmed[4:])
	// Strip trailing semicolon(s)
	rest = strings.TrimRight(rest, ";")
	rest = strings.TrimSpace(rest)
	// Strip surrounding double-quotes, preserving inner case
	if len(rest) >= 2 && rest[0] == '"' && rest[len(rest)-1] == '"' {
		rest = rest[1 : len(rest)-1]
	}
	return rest
}

// dispatch calls router.ProcessCommand, handles USE keyspace propagation,
// and prints the result. Returns (shouldExit, fatalError).
func dispatch(p *printer, sess *db.Session, sessMgr *session.Manager, stmt string) (bool, error) {
	// Detect USE statement before executing so we can propagate the keyspace
	// from the input rather than parsing the result string.  This is more
	// robust (handles quoted names like USE "MyKeyspace";) and consistent with
	// how the db executor parses the keyspace.
	pendingKs := parseUseKeyspace(stmt)

	result := router.ProcessCommand(stmt, sess, sessMgr)

	// Detect ExitSignal.
	if _, ok := result.(router.ExitSignal); ok {
		return true, nil
	}

	// On success, propagate the keyspace parsed from the input statement.
	// We only apply this when the result is not an error (a non-nil error result
	// means the keyspace didn't exist or the USE failed).
	if pendingKs != "" {
		if _, isErr := result.(error); !isErr {
			if err := sess.SetKeyspace(pendingKs); err != nil {
				fmt.Fprintf(os.Stderr, "Error switching keyspace: %v\n", err)
			} else {
				sessMgr.SetKeyspace(pendingKs)
			}
		}
	}

	_, _ = p.printResult(os.Stdout, result)
	return false, nil
}

// buildPrompt returns the primary prompt string.
// cqlai> when no keyspace, cqlai:<keyspace>> when one is set.
func buildPrompt(keyspace string) string {
	if keyspace == "" {
		return "cqlai> "
	}
	return fmt.Sprintf("cqlai:%s> ", keyspace)
}

// continuationPrompt returns a continuation prompt that aligns visually with
// the primary prompt (same width, ends with "... ").
func continuationPrompt(keyspace string) string {
	primary := buildPrompt(keyspace)
	width := len([]rune(primary))
	if width <= 4 {
		return "... "
	}
	return strings.Repeat(" ", width-4) + "... "
}
