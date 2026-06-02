# BDD Testing — Mandatory Guide

> **Audience:** every contributor — human and AI assistant.
> **Status:** binding. No exceptions without a written waiver in the PR body
> approved by a maintainer.

---

## TL;DR

1. **Every new feature, behaviour change, or bug fix MUST ship with at least one
   BDD scenario** (Gherkin `.feature` file + step definitions).
2. **PRs cannot be merged until the `bdd-tests` CI job is green.** The check is
   a required status on `main`.
3. Feature files live under `test/bdd/features/`. Step definitions live under
   `test/bdd/steps/` (or in-package under `internal/<pkg>/` when the function
   under test is unexported — see `internal/router/copy_options_bdd_test.go`).
4. Run locally: `go test -v -count=1 -run BDD ./test/bdd/... ./internal/router/...`
5. CI publishes every run's BDD artifacts (Gherkin output, JUnit, coverage) as
   GitHub Actions workflow artifacts on the `bdd-tests` job.

---

## The rule

**No BDD, no merge.**

A PR is mergeable only if **all** of the following are true:

| # | Requirement | Enforcement |
|---|---|---|
| 1 | At least one new `.feature` scenario covers the change | Reviewer + checklist |
| 2 | New scenarios cover happy path AND at least one invalid / edge case | Reviewer + checklist |
| 3 | All existing scenarios still pass | CI: `bdd-tests` job |
| 4 | All new scenarios pass | CI: `bdd-tests` job |
| 5 | New step definitions are reviewed (no `Skip`, no `// TODO`) | Reviewer |
| 6 | CHANGELOG entry references the new scenario in `Tests` or relevant section | `docs-quality-reviewer` |

### Exceptions (the only ones)

You may skip BDD only for these — and you must call it out explicitly in the
PR description under `### BDD waiver`:

- Pure formatting (`gofmt`, comment typos, file rename with no logic change).
- Pure dependency bump with no source change (e.g. `go.mod` only).
- CI / workflow changes that have no runtime effect on the binary.

If you are unsure, **write the test**. The cost of writing a Gherkin scenario
is lower than the cost of a regression that ships.

---

## Why BDD here

CQLAI is a CLI tool used interactively against live Cassandra clusters. A unit
test that says "function X returns Y" does not prove "user can run COPY FROM
PARQUET against a list<text> column without breaking." BDD scenarios encode
the second statement directly.

Concrete examples of what BDD has caught in this repo:

- `internal/router/copy_from_parquet_collection_test.go` — list vs set literal
  selection from destination schema (issue #81).
- `test/bdd/features/copy-options.feature` — silent option drops in COPY
  WITH-clause parser.
- `test/bdd/features/command-validation.feature` — destructive commands not
  guarded by confirmation.

---

## Where BDD code lives

```
test/
├── bdd/
│   ├── features/                # Gherkin .feature files (one per behaviour area)
│   │   ├── ai-command-parser.feature
│   │   ├── command-validation.feature
│   │   ├── copy-options.feature
│   │   ├── cql-splitter.feature
│   │   ├── save-command.feature
│   │   └── ssl-flags.feature
│   └── steps/                   # Step definitions (Go, godog)
│       ├── ai_steps_test.go
│       ├── save_steps_test.go
│       ├── splitter_steps_test.go
│       ├── ssl_flags_steps_test.go
│       └── validation_steps_test.go
└── integration/                 # +build integration — godog steps that need a live Cassandra
```

In-package BDD (for unexported functions): co-locate the step definitions
with the code under test:

```
internal/router/
├── copy_options_bdd_test.go     # godog suite calling parseCopyOptions directly
```

Use `_bdd_test.go` suffix so it is picked up by `go test` but kept distinct
from unit tests.

---

## Writing a scenario

### 1. Pick the right file

- New behaviour in an existing area → append to the existing `.feature`.
- New area → create `test/bdd/features/<kebab-name>.feature`.

### 2. Write the Gherkin first, before the code

This is non-negotiable. The scenario IS the spec. Acceptance criteria from the
issue map 1:1 to scenarios.

Skeleton:

```gherkin
Feature: <one-line capability statement>
  As a <user role>
  I want <observable behaviour>
  So that <business reason>

  Scenario: <one short, declarative sentence>
    Given <starting state>
    When <single action>
    Then <single observable outcome>
    And <optional further outcome>
```

Rules (enforced by `engineering-agents:bdd-guidelines`):

| Rule | Why |
|---|---|
| One `When` per scenario | A scenario describes one action |
| `Then` asserts observable behaviour, not implementation | Reviewers should not need to read source to grade the assertion |
| No persisted state between scenarios | Each scenario runs in isolation |
| Scenario titles use business language, not function names | Findable in failure output |
| Cover happy path AND at least one negative / edge case | Bugs hide on the edges |
| Use `Scenario Outline` + `Examples` for parameterised cases | DRY |
| No `Background` larger than 3 steps | Keeps scenarios self-contained |

### 3. Implement step definitions

```go
// test/bdd/steps/<area>_steps_test.go
package steps_test

import (
    "context"
    "testing"

    "github.com/cucumber/godog"
)

type myWorld struct {
    // per-scenario state — never package-level globals
}

func (w *myWorld) iDoX() error { /* ... */ }

func (w *myWorld) yIsObserved() error { /* ... */ }

func InitializeMyScenario(ctx *godog.ScenarioContext) {
    w := &myWorld{}
    ctx.Before(func(_ context.Context, _ *godog.Scenario) (context.Context, error) {
        *w = myWorld{} // reset
        return nil, nil
    })
    ctx.Step(`^I do X$`, w.iDoX)
    ctx.Step(`^Y is observed$`, w.yIsObserved)
}

func TestBDDMyArea(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeMyScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../features/<file>.feature"},
            TestingT: t,
        },
    }
    if suite.Run() != 0 {
        t.Fatal("BDD suite failed")
    }
}
```

Step rules:

- One `world` struct per suite. Reset in `Before`. No globals.
- Steps return `error`, not `t.Fatal` — godog formats errors nicely.
- Use real production code paths. **Never** mock the function you are
  validating. Mock only the boundary (network, filesystem) if needed.
- Regex captures: prefer named patterns in plain English. Avoid clever
  optional groups — duplicate the step instead.

### 4. Run locally

Pure (no Cassandra needed):

```bash
go test -v -count=1 -run BDD ./test/bdd/... ./internal/router/...
```

Single feature:

```bash
go test -v -count=1 -run TestBDDCopyOptions ./internal/router/
```

Integration (needs Cassandra on `127.0.0.1:9042`, e.g.
`docker run -p 9042:9042 cassandra:5.0`):

```bash
go test -v -tags integration -timeout 10m -count=1 ./test/integration/...
```

### 5. Verify CI

The `bdd-tests` job in `.github/workflows/ci.yml` runs on every push. It must
be green before merge. CI also runs the integration suite against the matrix
of Cassandra versions — your scenario must pass against **all** matrix
versions if it touches CQL behaviour.

---

## Accessing previous test runs

CI publishes BDD + integration artifacts (Gherkin reports, JUnit XML, raw `go
test` output, coverage profiles) as GitHub Actions workflow artifacts on the
`bdd-tests` job. Use these when triaging flakes, comparing runs, or auditing
what scenarios actually ran on a given commit.

Fetch via `gh`:

```bash
# List runs for a commit
gh run list --repo axonops/cqlai --commit <sha>

# Download all artifacts for a run
gh run download <run-id> --repo axonops/cqlai --dir ./ci-artifacts/

# Download artifacts for a PR's head commit
SHA=$(gh pr view 82 --repo axonops/cqlai --json headRefOid -q .headRefOid)
RUN=$(gh run list --repo axonops/cqlai --commit "$SHA" --json databaseId -q '.[0].databaseId')
gh run download "$RUN" --repo axonops/cqlai --dir ./ci-artifacts/
```

Retention is governed by the repo's GitHub Actions artifact retention policy.

### Pretty HTML report

The Cucumber JSON output is consumable by any Cucumber HTML reporter. After
downloading the artifact:

```bash
npx cucumber-html-reporter \
  --cucumber-json ./ci-artifacts/bdd-tests/godog-cucumber.json \
  --output report.html
open report.html
```

---

## Pre-merge checklist (copy into your PR description)

```markdown
### BDD compliance

- [ ] Added or updated `.feature` file(s): <path>
- [ ] Scenarios cover happy path AND at least one invalid / edge case
- [ ] Step definitions added or updated: <path>
- [ ] `go test -v -count=1 -run BDD ./test/bdd/... ./internal/router/...` passes locally
- [ ] CI `bdd-tests` job is green
- [ ] CHANGELOG entry mentions the new scenario(s)
- [ ] (If CQL-affecting) integration suite green on every Cassandra matrix version
```

If you set `### BDD waiver`, justify it in one paragraph and tag a maintainer.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `pending step` in godog output | Regex in `ctx.Step(...)` does not match the Gherkin sentence | Print godog's suggested snippet, paste the new step |
| Scenario passes locally, fails in CI | Hidden state leaks between scenarios | Reset `world` in `Before`. Never use package-level vars |
| `bdd-tests` job times out | Suite calls a real network in a step | Mock the network boundary only. Production code stays untouched |
| Cassandra integration scenario passes on 5.0, fails on 2.1 | CQL grammar / system table differs | Gate the scenario with a tag and version-skip in step setup |
| Want to add a step but it duplicates an existing one | You probably don't | Reuse the existing step. Rename for clarity if needed |

---

## For AI assistants

If you are an AI (Claude, Copilot, Cursor, etc.) implementing a feature in
this repo:

1. **Before writing implementation code, write the Gherkin.** The user can
   correct the spec faster than the implementation.
2. Always run `engineering-agents:bdd-guidelines` skill in parallel with the
   stack-specific skill (e.g. `python-bootstrap:pytest-bdd-tests`,
   `go-bootstrap:godog-bdd-tests`).
3. **Never** mark a task complete with `Skip` or `t.Fatal("TODO")` in a step.
   The CI gate will reject it.
4. If you cannot satisfy a scenario, stop and ask. Do not delete the scenario
   to make CI green.
5. When triaging a CI failure, download the workflow artifact first
   (`gh run download ...`), do not guess from the GitHub Actions log
   summary.

---

## References

- `engineering-agents:bdd-guidelines` skill — generic BDD rules (load first).
- `go-bootstrap:godog-bdd-tests` skill — Go/godog specifics.
- [Cucumber Gherkin reference](https://cucumber.io/docs/gherkin/reference/).
- [godog README](https://github.com/cucumber/godog).
- CHANGELOG.md — every PR appends a `Tests` entry referencing the new scenario.
