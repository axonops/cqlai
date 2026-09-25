# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.2] - 2026-09-25

Three things that were silently not working: a copy that went nowhere and said
nothing, a view that ate the keys for getting out of it, and no shortcut for
quitting at all.

### Added

- **`Ctrl+Q` quits**, and the FILE menu shows it beside `QUIT`. It asks first,
  the same as picking `QUIT` does — a shortcut printed next to an entry has to
  do what that entry does. `Ctrl+Q` on all three platforms, macOS included:
  `⌘+Q` never reaches a terminal application, because the terminal takes it and
  quits itself.

### Fixed

- **The shell's keys are no longer typed into the AI conversation.** In CHAT,
  `Alt+F` opened no FILE menu, `Alt+H` and `F1` opened no help, and the letters
  went into the message instead. The view handed every key it did not recognise
  to its input field, and the list of keys it let past named `F2` to `F6` and
  nothing else — a second copy of the main handler's list, which stopped being
  updated when `Alt+F` and `Alt+A` were added. There is one list now, and the
  view keys come from the tabs themselves.
- **Copying says when it had nowhere to go.** On a stock Linux desktop a
  drag-select copied nothing and said nothing: no VTE-based terminal supports
  OSC 52 — GNOME Terminal and Ptyxis are built on it — and `wl-clipboard`,
  `xclip` and `xsel` are none of them installed by default. Pasting back inside
  CQLAI went on working, because that falls back to what CQLAI itself last
  copied, so it looked as though the copy had worked. The first copy in a
  session that finds no clipboard tool now says so, and names the one to
  install.

  Nothing in a build can make that copy work — what is missing is the tool:
  `sudo apt install wl-clipboard`, or `xclip` on X11.

### Documentation

- The README gave a package name where it should have given a command, and said
  a selection "lands on the system clipboard when you let go" without saying
  what Linux needs first. `docs/INSTALLATION.md` gains a **Runtime
  Requirements** section, which is where someone looks before hitting this
  rather than after.
- Its prerequisites said Go 1.21, which cannot build a module declaring
  `go 1.26.6`.
- The README said "There is no right-click paste" two paragraphs after another
  section describing right-click paste in detail. There is one.
- Tests now fail when the README does not name a clipboard tool CQLAI tries,
  when a README or the in-app help does not name the quit key, and when the
  macOS column offers a key the terminal takes.

### Known limitation

The note about a copy having nowhere to go is written to the CONSOLE view,
which the RESULTS view does not draw — so copying query output, which is the
likeliest place to do it, will not show it. It belongs on the row above the
prompt, which every view draws.

## [0.2.1] - 2026-09-23

Copying with the mouse did not reach the system clipboard. On macOS nothing was
copied at all; on Windows the clipboard was cleared instead. Both looked like
they had worked, because pasting back into CQLAI falls back to what CQLAI
itself last copied, so the text never had to leave the process to seem fine.

### Fixed

- **macOS**: releasing a drag-selection wrote the text out through OSC 52 only.
  Terminal.app drops that write outright and iTerm2 refuses it until
  "Applications in terminal may access clipboard" is enabled, so the highlight
  appeared and the clipboard never changed. `Command + C` could not make up for
  it: mouse reporting is on by default and has already taken the terminal's own
  click-and-drag selection away. Copying now also tells the machine, through
  `pbcopy` — the same way the paste side already asks it with `pbpaste`
  ([#210](https://github.com/axonops/cqlai/issues/210)).
- **Windows**: the new write went out as `powershell.exe -NoProfile -Command
  Set-Clipboard`, which runs the cmdlet with no `-Value`. A process's stdin does
  not reach the PowerShell pipeline, so it cleared the clipboard rather than
  setting it — and exited 0, which the caller took for success and stopped
  trying. Copying in CQLAI silently destroyed whatever you had copied
  elsewhere. The text is piped in with `$input` now, and `clip.exe` follows as a
  second route for a machine where PowerShell is missing or refuses to run.

Over ssh with no clipboard tool, OSC 52 is still the only route that can work,
and it is still taken. Locally one of the two lands.

## [0.2.0] - 2026-09-18

Sixty merged changes since 0.1.7, most of them the terminal UI. CQLAI stopped
being a prompt that prints tables and became a shell you can work in with the
mouse, with the views on tabs, the file operations on a menu, the settings in a
window, and the AI reading what is already on screen rather than only writing
CQL.

### Added

- **Clickable tabs for the views**, with the key for each on the tab. `RESULTS`
  holds two tabs of its own — the last query's output, and the trace of the
  requests that fetched it. A view with nothing to show is dimmed rather than
  hidden.
- **SCHEMA browser** (`F3`): the cluster's keyspaces and tables in a tree, with
  each object's definition beside it. It notices a schema change wherever it
  was made, including from another window, and forgets what a DDL statement
  changed rather than showing a stale definition.
- **FILE menu** (`Alt+F`): `SOURCE`, `COPY TO`/`COPY FROM`, `SAVE RESULTS`,
  `AUTOSAVE`, `CONNECT`, `PREFERENCES` and `QUIT`, each with a form. Path
  fields complete as you type and open a file browser you can walk with the
  keyboard or the mouse.
- **PREFERENCES window**: every setting in `cqlai.json`, edited in the shell and
  written back to the file it was loaded from. `CHAT` holds which provider to
  ask; each provider's key, model and URL are in its own section.
- **CONNECT window with saved connections**: a list on the left, the settings on
  the right, a default connection at the top, and the one in use marked. CQLAI
  now starts without a cluster and connects from here.
- **Mouse throughout**: drag to select text and it lands on the system
  clipboard, right-click to paste, click the status line settings and the query
  info fields to change them, click the tabs, drag the line between panes to
  resize them, and scroll with the wheel over whichever pane the pointer is on.
- **Trace analysis with the AI** (`Alt+A` in the TRACE tab): where the time
  went, what the trace shows that the timings do not, and what to change —
  under the trace it is about, in a pane you can resize and copy out of.
- **Schema review with the AI** (`Alt+A` in the SCHEMA view): what a table is,
  what will go wrong with it, and what to change, including when a change means
  writing the data again somewhere else.
- **Tab completion for whole CQL statements**, including table properties,
  `CREATE` of every object, and DML. Where a name cannot be looked up it says
  what to type — `<table name>`, `<column name>`, `<value>` — rather than
  stopping with no suggestion.
- **Help window** (`F1` or `Alt+H`), scrollable, listing every command and key.
- **`AUTOSAVE`** — what `CAPTURE` was — now takes a directory and writes every
  query from then on.
- **Parquet output** from the Save window, with values converted from their
  displayed form.
- Scrollbars for the Console, the Command History list, and both trace panes.
- Key column markers in results, with the position of each in the primary key,
  and a line under each column name saying what it is.
- Gemini and Ollama can now be used for chat. Neither ever worked: Gemini had
  configuration and no code behind it, and Ollama read replies from its
  OpenAI-compatible endpoint in the shape of its own API, so every reply parsed
  as empty.

### Changed

- Migrated to **Charm v2** (`bubbletea`, `bubbles`, `lipgloss`).
- The default configuration path is now `~/.cassandra/cqlai.json`, beside
  `cqlshrc`.
- One set of AI provider clients. A second set behind an `AIClient` interface
  was unreachable, and routing one feature through it added 5.1MB to the binary
  — the reason ANTLR was removed from this project. OpenAI, OpenRouter, Gemini
  and Ollama are one conversation with four addresses.
- Switching `OUTPUT` redraws the result on screen from the values behind it,
  rather than leaving the previous format there.
- `ASCII` dropped from the Save formats.
- The AI tab is left out entirely when no provider is configured.

### Fixed

- Clicking the header row crashed CQLAI.
- The Save window wrote no file, for any format.
- Collection columns were saved as `null` rather than as text.
- `ReadBatch` discarded every row past the first batch.
- `NULL`s were lost when scanning rows.
- `COPY FROM PARQUET` formatted values into the statement instead of binding
  them, in both the single-file and partitioned readers.
- `Close` reported an error on success in the Parquet writer.
- `DESCRIBE` ordered primary key columns arbitrarily rather than by schema
  position.
- Paging did not reach the end of `EXPAND` output.
- A trace captured only the first page of a paged query, and reported that
  page's row count as the query's.
- `go install` failed on a self-replace directive.
- The AI was handed an empty database by the schema context.
- Row counts, sideways scrolling and `PAGING`; the status and info bars now fit
  the terminal; `scrollInfo` no longer draws past the right edge; the command
  history closes when you press outside it; the trace scrolls sideways.
- The destructive-operation warning above a generated `DROP` or `DELETE` is now
  actually set — it was written by a validator only the unreachable half of the
  AI package called.
- `temperature` is no longer sent to Anthropic models that reject it.

### Security

- All reachable vulnerabilities cleared; `arrow-go` upgraded to `v18.7.0`.

### Build

- Leftover ANTLR references removed from the build.
- Documentation is checked against the code that decides it: the views and
  their keys, the FILE menu, and the configuration paths all fail the build
  when the README and the code disagree.

## [0.1.7] - 2026-06-02

### Added

- BDD test suites (godog) covering pure-logic packages: `validation` command
  syntax + dangerous-command guard, `batch` CQL statement splitter and shell
  command detection, `router` SAVE-command parser, `ai` JSON command parser,
  and `router` COPY WITH-clause option parser. 116 new scenarios.
- CI job `bdd-tests` running all godog suites on every push.
- Cassandra 5.0 integration step in `cassandra-tests` job — runs the Go
  `+build integration` suite (`go test -tags integration ./test/integration/...`)
  against the live Cassandra 5.0 service container.

### Changed

- Repo-wide `gofmt` pass — formatting only, no behavioural changes.

### Fixed

- `COPY FROM PARQUET` no longer emits set literals (`{...}`) for `list<...>`
  columns whose names happen to match the old heuristic (`tags`, `*_set`,
  `*_nums`, anything containing `unique`). Collection brackets are now chosen
  from the destination table's `system_schema.columns.type` — both for the
  single-file and partitioned readers. Unblocks `TestRoundTripCollections`
  on Cassandra 5.0; CI `-skip` flag removed
  ([#81](https://github.com/axonops/cqlai/issues/81)).

### Security

- Bump transitive `github.com/apache/thrift` from `v0.22.0` to `v0.23.0` to
  resolve [CVE-2026-41602](https://nvd.nist.gov/vuln/detail/CVE-2026-41602)
  (`TFramedTransport` integer-overflow, GHSA-wf45-q9ch-q8gh, CVSS 7.5).

### Build

- Bump Go toolchain from `1.26.1` to `1.26.3` across `go.mod`, all GitHub
  Actions workflows, and Docker build images.
