# 🛠 Roadmap & Milestones — cqlsh-js

## Phase 1 – MVP (Barebones REPL)
**Goal:** Connect and query.  
- Connect via `cassandra-driver`.  
- Simple REPL (`query -> output`).  
- `DESCRIBE KEYSPACES` via metadata.  
- Print errors cleanly.  
- Startup banner.  

---

## Phase 2 – Core Shell Parity
**Goal:** Match essential `cqlsh`.  
- Command router.  
- `DESCRIBE` (tables, UDTs, functions).  
- `CONSISTENCY`, `PAGING`, `TRACING`.  
- Paging support with `fetchSize`.  
- Query history.  
- Status bar.  
- Config (`.cqlshrc-js`, env vars).  

---

## Phase 3 – Modern UI Polish
**Goal:** Sexy Ink-powered UI.  
- Ink layout: results pane + sticky input + footer.  
- Zebra-striped tables, capped widths.  
- Syntax-highlighted input.  
- Splash screen (gradient).  
- Toast notifications.  
- Spinner + progress for queries.  
- Color themes (accent, warn, err).  

---

## Phase 4 – Advanced Features
**Goal:** Implement heavier functionality.  
- `COPY TO/FROM` with CSV streaming.  
- Tracing integration (`system_traces`).  
- Multi-contact point support.  
- Keybindings:  
  - `F2` tracing toggle.  
  - `F3` cycle consistency.  
  - `F4` set paging size.  
  - `PgUp/PgDn` result navigation.  
- Theme toggle (dark/light).  
- Richer error diagnostics.  

---

## Phase 5 – Stretch Goals
**Goal:** Beyond cqlsh.  
- Schema browser side panel.  
- Unicode sparklines for latency.  
- Saved connection profiles.  
- Multi-line editor.  
- Plugin/hook system.  

---

## Suggested Timeline
- **Phase 1:** 1 week  
- **Phase 2:** 2–3 weeks  
- **Phase 3:** 2 weeks  
- **Phase 4:** 3–4 weeks  
- **Phase 5:** ongoing / optional  



---

## 🔹 Add to `ROADMAP.md`

```markdown
## Phase NLQ-A — Schema-Aware, Read-Only NLQ
**Goal:** First usable NLQ integration.  
- Build schema snapshot system.  
- Add provider abstraction + prompt templates.  
- Generate CQL from English + rationale.  
- Implement validator (partition key, columns, limit).  
- Ink UI: NLQ tab with query, rationale, and flags.  
- Hard block on writes.

---

## Phase NLQ-B — Quality & Ergonomics
**Goal:** Improve reliability and usability.  
- Auto-repair queries on validation failure.  
- Map natural time ranges → clustering key filters.  
- Toggle for `ALLOW FILTERING` with warning.  
- Suggest “common queries” per table (based on schema).

---

## Phase NLQ-C — Power-Ups
**Goal:** Go beyond basic query gen.  
- Extract parameters from English into prepared statements.  
- Learn from user edits (few-shot fine-tuning).  
- (Optional) Write intents with two-step confirmation & dry-run support.

---

## NLQ Timeline
- **Phase NLQ-A:** 1–2 weeks  
- **Phase NLQ-B:** 1–2 weeks  
- **Phase NLQ-C:** optional / later  
