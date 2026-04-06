---
name: dlpbuster-dev
description: >
  Core development skill for building the dlpbuster CLI tool — an open-source Go-based
  automated DLP bypass tester for red teamers. Use this skill whenever the agent is
  working on any part of the dlpbuster project: implementing exfil channel modules (DNS,
  HTTPS, ICMP, S3, GCS, Azure, SMTP, webhook), building the core engine, writing tests,
  extending the CLI, building the callback listener, generating reports, or scaffolding
  new Go files. Triggers on any task involving Go code, channel modules, the Channel
  interface, the engine runner, payload generation, result aggregation, or CLI commands
  within this project. Also triggers for project maintenance tasks: updating TODO.md,
  committing to GitHub, running tests, linting, or building releases.
---

# dlpbuster Dev Skill

This skill governs all development work on the `dlpbuster` project. Read this file at the start of every session before writing any code.

## Project Snapshot

| Property | Value |
|---|---|
| Tool | `dlpbuster` — automated DLP bypass tester |
| Language | Go 1.22+ |
| Binary | Single static binary, no CGO (except ICMP) |
| CLI framework | `cobra` + `viper` |
| UI | `bubbletea` + `lipgloss` |
| Logging | `zerolog` |
| Testing | `testify` |
| Build | `goreleaser` |
| Target users | Red teamers, pentesters |
| License | MIT |

---

## Session Startup Checklist

Run these steps at the start of every session in order:

```bash
# 1. Confirm working directory
pwd   # must be /workspace/dlpbuster

# 2. Check current state
git status
git log --oneline -5

# 3. Read TODO.md to find next open task
cat TODO.md | grep "^\- \[ \]" | head -20

# 4. Run tests to confirm baseline is green
go test ./... -race -timeout 30s

# 5. Run linter
golangci-lint run ./...
```

Do NOT start writing code until step 4 passes. If tests fail on a clean checkout, file a bug in memory before touching anything.

---

## Core Architecture (memorize this)

### Channel Interface — never change this signature

```go
// internal/channels/channel.go
type Channel interface {
    Name()        string
    Description() string
    Run(ctx context.Context, cfg ChannelConfig) Result
}
```

### Result struct — the unit of truth

```go
type Result struct {
    Channel   string
    Status    Status        // Passed | Blocked | Partial | Error
    BytesSent int
    Duration  time.Duration
    Evidence  []string      // raw log lines, HTTP status codes, DNS responses
    Error     error
}
```

### Engine flow

```
config.Config
    │
    ├── payload.Generator  →  []byte (encrypted, compressed, chunked)
    │
    └── engine.Run(channels, payload)
            │
            ├── dns.Channel.Run()    → Result
            ├── https.Channel.Run()  → Result
            └── ...
            │
            ▼
     engine.Aggregate([]Result)  →  report.Render()
```

---

## Implementing a New Channel

Follow this exact sequence every time:

### Step 1 — Create the module file

```
internal/channels/<name>/<name>.go
```

```go
package <name>

import (
    "context"
    "github.com/your-org/dlpbuster/internal/channels"
)

// <Name>Channel exfiltrates payload via <protocol>.
// Requires: <any privilege or network requirements>.
type <Name>Channel struct{}

func New() *<Name>Channel { return &<Name>Channel{} }

func (c *<Name>Channel) Name() string        { return "<name>" }
func (c *<Name>Channel) Description() string { return "<one sentence>" }

func (c *<Name>Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
    start := time.Now()
    // implementation
    return channels.Result{
        Channel:   c.Name(),
        Status:    channels.Passed,
        BytesSent: len(cfg.Payload),
        Duration:  time.Since(start),
        Evidence:  []string{"..."},
    }
}
```

### Step 2 — Register in registry

```go
// internal/channels/registry.go
import "<name>" "github.com/your-org/dlpbuster/internal/channels/<name>"

func All() []Channel {
    return []Channel{
        dns.New(),
        https.New(),
        <name>.New(),   // add here
    }
}
```

### Step 3 — Write the test

```go
// internal/channels/<name>/<name>_test.go
func TestRun_Blocked(t *testing.T) {
    t.Parallel()
    // use mock listener — no real network calls
}

func TestRun_Passed(t *testing.T) {
    t.Parallel()
    // confirm Result.Status == Passed and BytesSent > 0
}
```

### Step 4 — Document

```
docs/channels/<name>.md
```

Include: protocol description, config options, privilege requirements, example output, known detection signatures.

### Step 5 — Mark TODO.md

Use the filesystem MCP tool to tick the checkbox in TODO.md.

---

## Testing Rules (enforce strictly)

```bash
# Always run with -race
go test ./... -race -timeout 30s

# For a single package
go test ./internal/channels/dns/... -race -v

# Coverage check
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1
# Target: >80% on internal/ packages
```

**Mock all I/O:**
- HTTP: use `net/http/httptest.NewServer`
- DNS: use a fake resolver (stub `net.Resolver`)
- Cloud (S3/GCS/Azure): use the provider's official mock client or `httptest`
- Never hit real external endpoints in tests
- Tests must pass completely offline

**Table-driven tests preferred:**

```go
tests := []struct {
    name   string
    input  channels.ChannelConfig
    want   channels.Status
}{
    {"payload delivered", cfg, channels.Passed},
    {"timeout", timeoutCfg, channels.Blocked},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        got := ch.Run(ctx, tt.input)
        assert.Equal(t, tt.want, got.Status)
    })
}
```

---

## CLI Command Rules

Every subcommand lives in `cmd/dlpbuster/<command>.go`.

```go
// Standard command structure
var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run exfil channel tests",
    RunE:  runChannels,
}

func runChannels(cmd *cobra.Command, args []string) error {
    // 1. Load config via config.Load()
    // 2. Build payload via payload.Generate(cfg)
    // 3. Select channels via registry.Filter(cfg)
    // 4. Run engine.Run(ctx, channels, payload, cfg)
    // 5. Render via report.Render(results, format)
    return nil
}
```

Shared setup (logging init, config load, ethics banner) goes in `PersistentPreRunE` on the root command — never duplicated per subcommand.

---

## Security Constraints (non-negotiable)

These rules are enforced in code, not just policy:

1. **Synthetic payloads only by default.** `payload.Generate()` produces random bytes. `--payload-file` flag is the only way to use real data — and it must print a warning.

2. **Ethics banner always prints** — hardcoded in `PersistentPreRunE`, cannot be suppressed even with `--silent`.

3. **Raw sockets check elevation first:**

```go
if os.Getuid() != 0 {
    return channels.Result{
        Status: channels.Error,
        Error:  errors.New("ICMP channel requires root (run with sudo)"),
    }
}
```

4. **All runs log locally** to `~/.dlpbuster/logs/<timestamp>.log`. Log write failures are non-fatal but must be warned.

5. **No persistence, no C2, no credential harvesting** — any PR adding these is immediately rejected.

---

## Payload Package Reference

```go
// Generate a synthetic payload
p, err := payload.Generate(payload.Options{
    SizeBytes: cfg.Payload.SizeBytes,   // default 1024
    Encrypt:   cfg.Payload.Encrypt,     // AES-256-GCM
    Compress:  cfg.Payload.Compress,    // gzip
})

// Split for chunked channels
chunks := payload.Split(p, chunkSize)

// From file (red-flag warning printed automatically)
p, err := payload.FromFile("/path/to/file", opts)
```

---

## Report Package Reference

```go
r := report.New(results, report.Options{
    Format:     report.Markdown,   // Human | JSON | Markdown | HTML
    TargetName: "acme-corp",
    RunAt:      time.Now(),
})

// Write to stdout
r.Render(os.Stdout)

// Write to file
r.WriteFile("./dlpbuster-report.md")
```

Reports always include: channel status table, bytes sent, duration, evidence log lines, pass/block/partial summary, risk narrative.

---

## Commit Convention

```
feat(dns): add TXT record exfil mode
fix(engine): handle nil result from timed-out channel
test(https): add chunked payload unit test  
docs(icmp): document raw socket privilege requirement
refactor(payload): extract encrypt/compress into sub-packages
chore(ci): add goreleaser release workflow
```

Scope must be one of: `dns`, `https`, `icmp`, `s3`, `gcs`, `azure`, `smtp`, `webhook`, `engine`, `payload`, `listener`, `report`, `config`, `ui`, `cli`, `ci`, `docs`.

**Branch per task:**
```bash
git checkout -b feat/dns-txt-record
# implement
git commit -m "feat(dns): add TXT record exfil mode"
git push origin feat/dns-txt-record
# open PR via GitHub MCP tool
```

Never commit directly to `main`.

---

## Make Targets

```bash
make build      # go build → ./bin/dlpbuster
make test       # go test ./... -race -timeout 30s
make lint       # golangci-lint run ./...
make coverage   # generate coverage.html
make release    # goreleaser --snapshot --clean
make clean      # remove ./bin and coverage artifacts
```

Always run `make test && make lint` before opening a PR.

---

## Common Mistakes to Avoid

| Mistake | Correct approach |
|---|---|
| Hitting real network in tests | Use `httptest.NewServer` or fake resolver |
| Global vars for config/deps | Inject via struct fields or context |
| `panic()` in library code | Always return `error` |
| Unicode bullets in output | Use lipgloss styles only |
| Hardcoding endpoints | Always read from `config.Config` |
| Functions over 50 lines | Extract helpers immediately |
| Skipping `-race` flag | Always test with `-race` |
| Missing godoc on exports | Every exported symbol needs a comment |
| Committing to `main` | Always use a feature branch |

---

## Reference Files

Read these when working on the relevant area:

- `agent-instructions.md` — Full coding standards, ethics rules, scope boundaries
- `TODO.md` — Current task queue, tick checkboxes as you complete items
- `project-structure.md` — Full directory tree and package responsibilities
- `docs/channels/<n>.md` — Per-channel setup and protocol details
- `docs/usage.md` — Full CLI usage reference
- `mcp.json` — Available MCP servers and their capabilities
