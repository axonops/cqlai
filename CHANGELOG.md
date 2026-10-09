# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.4] - 2026-10-09

### Added

- **TLS for the MCP server.** `cqlai mcp` serves HTTPS, at TLS 1.2 or
  later, when given a certificate and key. With a client CA, every client
  has to show a certificate that CA signed. They are set in PREFERENCES under
  MCP SERVER (`TLS certificate`, `TLS key`, `TLS client CA`), in the `mcp`
  block of `cqlai.json` (`tlsCert`, `tlsKey`, `tlsClientCA`), or with
  `--tls-cert`, `--tls-key` and `--tls-client-ca`.
- **The address the MCP server listens on.** `Listen on` in PREFERENCES,
  `listen` in `cqlai.json`, or `--listen`. It is `127.0.0.1` unless set. On
  any other address the network can reach the server, so it then needs TLS,
  and the token or a client CA. Without them it does not start, and
  `FILE > MCP SERVER` says why.

### Changed

- **The MCP token is off unless turned on.** Turn it on with `Require token`
  in PREFERENCES, `"token": true` in `cqlai.json`, or `cqlai mcp --token`.
  With it off, the client configuration has no `Authorization` header. A
  request from a web page is still refused, with the token on or off.
- **The MCP token and audit log are in `~/.cassandra`,** beside `cqlai.json`:
  `~/.cassandra/cqlai_mcp_token` and `~/.cassandra/cqlai_mcp_audit.log`. They
  were hidden files in the home directory. An old file is moved the first time
  it is needed, so a client set up with the old token keeps working.

## [0.3.3] - 2026-10-09

### Security

- **Built with Go 1.27.2 and `golang.org/x/net` v0.60.0.** v0.3.2 had 12
  known vulnerabilities. They are in Go's `net/http`, `net/textproto`,
  `crypto/tls` and `html/template`, and in the HTTP/2 code in `x/net`. The
  `net/http` server ones apply while `cqlai mcp` is serving. govulncheck
  finds none in this release.
- **Hidden keyspaces stay hidden from an MCP client when every keyspace is
  allowed.** With a deny list and no keyspace list, the `query` tool could
  read the system keyspaces. Through them it could see the names of hidden
  keyspaces, tables and columns, their sizes, and the statements traced
  against them. Now, when every keyspace is allowed, `system` and every
  `system_` keyspace are hidden as well. A keyspace list can still name one.
  `node_status` still reads `system_views`, unless it is denied or left out
  of a keyspace list. The CHAT view is unchanged.

### Changed

- **Every dependency is at its latest version.** Among them are
  `anthropic-sdk-go` v1.79.1, `bubbletea` v2.1.0 and `arrow-go` v18.8.0.
- **Building from source needs Go 1.27.**

### Fixed

- **`DESCRIBE FUNCTION` with a quote in the name.** The name is now written
  into the lookup as a quoted CQL string.
- **The APT and YUM install instructions.** The repository key, suite and
  YUM address they gave did not exist. They now match the repositories.

## [0.3.2] - 2026-10-08

### Added

- **Release binaries and tarballs for Linux and macOS.** Each release now has
  `cqlai-<version>-<os>-<arch>.tar.gz` for `linux-amd64`, `linux-arm64`,
  `darwin-amd64` and `darwin-arm64`. The binary inside is called `cqlai`. The
  binaries on their own are `cqlai-linux-amd64`, `cqlai-linux-arm64`,
  `cqlai-darwin-amd64` and `cqlai-darwin-arm64`, so
  `releases/latest/download/<name>` always gets the newest. The macOS
  binary is signed and notarized. Its checksums are in `SHA256SUMS-macos.txt`.
- **Pick a completion with the mouse.** A click on an entry in the
  completion list uses it, the same as Enter. The wheel moves through the list.

### Changed

- **The Linux tarballs are renamed.** `cqlai-linux-amd64.tar.gz` is now
  `cqlai-0.3.2-linux-amd64.tar.gz`, and the binary inside is `cqlai` rather
  than `cqlai-linux-amd64`.

### Fixed

- **Tab in a comma-separated PREFERENCES setting keeps the names before it.**
  It used to replace the whole list with the name picked.
- **Esc works in every dialog.** It did nothing in some of them, such as
  "No AI provider is configured" from the trace analysis.
- **Ctrl+U deletes all of a line of Japanese or other wide text.** It
  deleted half of it. Ctrl+K, Ctrl+W, Ctrl+Y, Alt+D and Ctrl+Left/Right had
  the same problem.
- **Option+F, Option+H, Option+D and Option+B work on a Mac** without
  changing the terminal's settings. macOS types `ƒ`, `˙`, `∂` and `∫` for
  them, and those are now read as the shortcuts. The README gives the
  setting that makes Option send Alt for the rest.
- **Esc in the console no longer asks to quit.** It also cleared what was
  typed.
- **The completion list no longer covers the prompt,** and its key help no
  longer wraps onto a second line.

## [0.3.1] - 2026-10-07

### Fixed

- **`TRACING ON` works in the shell while an MCP client is connected.** With
  `cqlai mcp` running, a client that called `describe` before anything was
  typed in the shell took over the shell's settings: `TRACING ON`, typed or
  picked on the status bar, turned on tracing for the model's session, and the
  status bar stayed `Trace: OFF`. `CONSISTENCY`, `PAGING`, `OUTPUT` and
  `AUTOFETCH` went the same way. The MCP server now runs its `DESCRIBE`
  statements apart from the shell's.

## [0.3.0] - 2026-10-07

### Added

- **An MCP server.** An AI assistant can work with one cluster through CQLAI,
  over the Model Context Protocol. There are two ways to run it:
  - `cqlai mcp` opens the shell and serves MCP at `http://127.0.0.1:7845/mcp`,
    for whichever connection is picked in `FILE > CONNECT`. Every request has
    to carry the token kept in `~/.cqlai_mcp_token`, and requests from a web
    page are refused. `FILE > MCP SERVER` prints a client configuration ready
    to copy, and the status bar shows `MCP: :7845`.
  - `cqlai mcp --headless` serves it on stdin and stdout, for a client that
    starts CQLAI itself.

  Neither needs a cluster to start: a tool called without one says so, and the
  next call tries again.

  The tools: `connection_info`, `list_keyspaces`, `list_tables`, `get_schema`,
  `fuzzy_search`, `describe`, `query` (a page at a time), `trace_query`,
  `node_status`, `table_size`, `list_roles` and `propose_change`. Keyspace and
  table definitions are offered as resources, and the client is told when the
  schema changes. Two prompts, `review_table` and `diagnose_query`, use the
  same instructions as `Alt+A`.
- **The MCP server never changes anything.** A model that wants an `INSERT`,
  `UPDATE`, `DELETE`, `BATCH`, `CREATE`, `ALTER`, `DROP` or `TRUNCATE` proposes
  it with `propose_change`, and gets back the statement and what it will do -
  warnings written into CQLAI for each kind of statement, not the model's own
  account. In `cqlai mcp` the statement is also put in the prompt, unrun, and
  running it is always confirmed.
- **What the model may do is set in `PREFERENCES` and `CONNECT`**, under
  `MCP SERVER`: which of `SELECT`, `DESCRIBE` and `LIST` it may run, which
  keyspaces it can see, tables it can never see, columns whose values are
  hidden, whether scans are allowed, and the limits. A connection's settings can
  only narrow the ones in `PREFERENCES`. In the shell they apply as soon as they
  are saved.

  Every statement passes one gate: one statement per call, a permitted command,
  tables named as `keyspace.table`, nothing hidden, and no `ALLOW FILTERING` or
  cross-partition aggregate unless scans are allowed. `system_auth` is always
  hidden. A redacted column's values cannot come back by alias, `JSON`, a
  function or a guess in `WHERE`. What is refused comes back as the statement,
  for the user to run if they choose. Every call is written to
  `~/.cqlai_mcp_audit.log`, without values. The server keeps a session of its
  own, apart from the shell's.
- **The schema tree can be filtered.** The heading above the `SCHEMA` tree is a
  filter: Up from the top row, `/` with nothing at the prompt, or a click gives
  it the keys. The tree narrows to the keyspaces and tables whose names contain
  what is typed, a table shown under its keyspace without opening it. `Esc`
  clears it, with what was found still selected.

### Changed

- **`fuzzy_search` finds keyspaces, and names by their sound.** It searched
  table names only, and only as spelled. It now searches keyspaces too, and
  finds names that sound alike (`hyto` finds `hayato`), have the search's
  letters in order, or are a typo away, saying why each matched. The `CHAT`
  view uses the same search.
- **The `CHAT` view's tools see the virtual keyspaces.**
- **The shell confirms a `DELETE` inside a `BATCH`, or a `DROP` behind a
  comment**, as it does one on its own: the check reads the statement rather
  than its first word.

### Fixed

- **A Console message of several lines is no longer double-spaced** when one of
  its lines is wider than the Console.

### Known issues

- The shell asks before a dangerous statement only when `requireConfirmation`
  is in the configuration file, or a connection flag is given. A statement
  proposed through MCP is always confirmed.

## [0.2.6] - 2026-10-05

### Fixed

- **Virtual keyspaces and tables are visible everywhere.** Cassandra 4.0 added
  tables the node computes rather than stores - `system_views.clients`,
  `.settings`, `.thread_pools` and forty-odd more. They are listed in
  `system_virtual_schema`, not `system_schema`, and every question CQLAI asked
  about what exists was asked of `system_schema`: a `SELECT` from
  `system_views.clients` worked, and `USE system_views` said the keyspace did
  not exist. They now appear in `USE`, `DESCRIBE KEYSPACES`, `DESCRIBE KEYSPACE`,
  `DESCRIBE TABLES`, the schema browser, the keyspace chooser and completion.
  `DESCRIBE TABLES` shows each one's key - the one thing a virtual table can be
  filtered on - and `virtual` where a stored table shows its compaction.
  `DESCRIBE SCHEMA` still leaves them out: it is meant to be replayed on another
  cluster, and a virtual table cannot be created with CQL.

  Before 4.0 there is no `system_virtual_schema`, and nothing changes.
- **`WHERE` completes the table's columns.** At a column after `WHERE`, `AND` or
  `SET`, completion offered the note `<column name>` rather than the columns -
  for every table, stored or virtual - because the walk over the statement had
  no way to ask what the columns were. It asks now.

## [0.2.5] - 2026-10-03

### Fixed

- **A definition line wider than the pane is copied whole.** Selecting a line
  that spills past the right edge - a long `PRIMARY KEY`, a compaction map -
  copied only the part of it that was showing.

  Two causes. The pane bound clamped the right edge, and that bound is there to
  keep a selection out of the tree, which is to the *left*: nothing is to the
  right but the edge of the screen, so the definition is unbounded on that side
  now. And that alone changed nothing, because for a drag along one line the
  end column is wherever the pointer was - and the end of a line truncated on
  screen cannot be pointed at. A drag held against the right-hand edge runs to
  the end of the line.

  The tree keeps its bound on both sides: the divider and the definition are to
  its right and neither belongs in a selection of it.

### Known limitation

The definition pane has no horizontal scrolling, so the rest of a wide line can
be copied but not read.

## [0.2.4] - 2026-10-02

    [ Review Schema Alt+A ]        [ Analyse Trace Alt+A ]

### Added

- **The AI buttons name the key that works them.** `Alt+A` had been bound since
  the feature shipped and was nowhere written down in the schema browser: the
  trace view had a line of text beside its button saying so, and the definition
  pane has no room for one, so nothing on screen said the button had a key at
  all - which is indistinguishable from not having one. Both buttons carry it
  now, and the trace's separate hint is gone, since with the key on the button
  it said it twice on one row.
- **`Alt+A` in the help inside the app**, which did not list it.

### Changed

- **A button that cannot run says so over the view, not in the console.**
  Picking `Review Schema` with no AI provider configured wrote a line into the
  console and switched to it, so being told why the button did nothing cost you
  the schema view and your place in it. It is a window over the view now, with
  one button, dismissed by Enter or Esc. The same for `Analyse Trace`, and for
  either button with nothing to work on.

### Documentation

- The README prose, both diagrams and the feature list show the buttons as they
  are drawn, key and all. A test asserts it, so the key cannot be taken off one
  button and left on the other.

## [0.2.3] - 2026-10-02

Selecting a definition out of the schema browser: it took the tree with it, it
crashed if you dragged off the bottom, and it could not reach past the screen.
All three were the same defect - a figure or a rule written in two places, with
only one copy maintained.

### Fixed

- **A selection stays in the pane it started in.** The view is two panes side
  by side, so a line of it is `tree │ definition`, and the rule every flowing
  selection uses takes the lines between its ends whole. Dragging over a
  definition to copy it copied the keyspace and table names beside it. The copy
  and the highlight each worked the span out separately, so that is one
  function now and what is painted is what is copied.
- **Dragging a selection off the bottom no longer crashes.** Dragging past an
  edge scrolls the content, and the guard against a view with no viewport to
  scroll was written on the top edge only - so the bottom edge dereferenced nil
  in the two views that draw as blocks rather than through a viewport, the
  schema browser and the trace. One rule for both edges.

  ```
  panic: runtime error: invalid memory address or nil pointer dereference
  ui.(*MainModel).extendSelection(...) selection.go:281
  ```
- **A definition taller than its pane can be selected whole.** The drag stopped
  at the bottom of the screen, so a long `CREATE TABLE` could only be copied a
  screenful at a time. The span is over the definition now rather than over the
  rows that happened to be painted, with the pane's scroll offset - the shape a
  viewport-backed view already has, where the scroll moves the text and the
  span together. So the drag scrolls, and the whole definition comes back
  ([#219](https://github.com/axonops/cqlai/issues/219)).

The tree deliberately does not scroll under a drag: it is a list of rows picked
with the keyboard, not a document to drag through.

### Documentation

- The README described 0.1.x. It now mentions the schema browser, saved
  connections, `PREFERENCES`, the `FILE` menu, statement-wide tab completion,
  and that the AI reads a trace and reviews a table rather than only writing
  CQL. `Coming Soon` is gone - what was on it has shipped.
- Three things in the AI section were wrong rather than missing: Anthropic was
  described as Claude 3, which is retired; "Queries are validated against your
  current schema" named a validator that carried `Schema: nil` and was deleted
  in 0.2.0; and nothing said the AI can be pointed at a trace or a definition.
- `docs/INSTALLATION_jp.md` gained the Runtime Requirements section its English
  counterpart has, and the Go 1.21 prerequisite that cannot build a module
  declaring `go 1.26.6`.

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
