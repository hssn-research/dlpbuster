# Project Structure — dlpbuster

## Overview
`dlpbuster` is a Go CLI tool. The structure follows standard Go project layout conventions with clear separation between public API, internal packages, CLI commands, and channel modules.

---

```
dlpbuster/
│
├── cmd/
│   └── dlpbuster/
│       ├── main.go               # Entry point — wires cobra root command
│       ├── run.go                # `dlpbuster run` — execute channels
│       ├── serve.go              # `dlpbuster serve` — start callback listener
│       ├── list.go               # `dlpbuster list` — print available channels
│       ├── report.go             # `dlpbuster report` — render last run report
│       └── config.go             # `dlpbuster config init` — setup wizard
│
├── internal/
│   │
│   ├── engine/
│   │   ├── engine.go             # Core runner: parallel channel execution, timeout
│   │   ├── engine_test.go
│   │   ├── result.go             # Result, Status types
│   │   └── aggregator.go         # Aggregate results, compute summary
│   │
│   ├── channels/
│   │   ├── channel.go            # Channel interface definition
│   │   ├── registry.go           # Channel registry (compile-time registration)
│   │   │
│   │   ├── dns/
│   │   │   ├── dns.go            # DNS tunnel exfil (A, TXT, MX records)
│   │   │   └── dns_test.go
│   │   │
│   │   ├── https/
│   │   │   ├── https.go          # HTTPS covert POST exfil
│   │   │   └── https_test.go
│   │   │
│   │   ├── icmp/
│   │   │   ├── icmp.go           # ICMP echo payload exfil (raw socket)
│   │   │   └── icmp_test.go
│   │   │
│   │   ├── s3/
│   │   │   ├── s3.go             # AWS S3 PUT exfil simulation
│   │   │   └── s3_test.go
│   │   │
│   │   ├── gcs/
│   │   │   ├── gcs.go            # GCP Cloud Storage upload
│   │   │   └── gcs_test.go
│   │   │
│   │   ├── azure/
│   │   │   ├── azure.go          # Azure Blob Storage upload
│   │   │   └── azure_test.go
│   │   │
│   │   ├── smtp/
│   │   │   ├── smtp.go           # Email attachment + body exfil
│   │   │   └── smtp_test.go
│   │   │
│   │   └── webhook/
│   │       ├── webhook.go        # Slack / Discord / Teams webhook exfil
│   │       └── webhook_test.go
│   │
│   ├── payload/
│   │   ├── payload.go            # Payload generator (random, file, pattern)
│   │   ├── encrypt.go            # AES-256-GCM encryption wrapper
│   │   ├── compress.go           # gzip compression
│   │   └── split.go              # Chunk payload into parts
│   │
│   ├── listener/
│   │   ├── listener.go           # Listener interface
│   │   ├── dns_listener.go       # DNS server (receives tunnel queries)
│   │   ├── https_listener.go     # HTTPS server (receives POST callbacks)
│   │   └── listener_test.go
│   │
│   ├── report/
│   │   ├── report.go             # Report builder
│   │   ├── markdown.go           # Markdown renderer
│   │   ├── html.go               # Self-contained HTML renderer
│   │   ├── json.go               # JSON renderer
│   │   └── report_test.go
│   │
│   ├── config/
│   │   ├── config.go             # Config struct + loader (viper)
│   │   ├── defaults.go           # Default values
│   │   └── config_test.go
│   │
│   └── ui/
│       ├── banner.go             # Startup banner + ethics warning
│       ├── progress.go           # Bubbletea progress / spinner
│       └── table.go              # Lipgloss results table renderer
│
├── docs/
│   ├── channels/
│   │   ├── dns.md
│   │   ├── https.md
│   │   ├── icmp.md
│   │   ├── s3.md
│   │   ├── smtp.md
│   │   └── webhook.md
│   ├── usage.md                  # Full CLI usage reference
│   └── setup-listener.md         # How to run the callback server
│
├── scripts/
│   ├── install.sh                # One-line installer
│   └── test-integration.sh       # Full integration test (requires network)
│
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                # Lint → test → build on push/PR
│   │   └── release.yml           # Goreleaser on tag push
│   └── ISSUE_TEMPLATE/
│       ├── bug_report.md
│       ├── feature_request.md
│       └── new_channel.md
│
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yml
├── .golangci.yml                 # Linter config
├── agent-instructions.md
├── mcp.json
├── TODO.md
├── project-structure.md          # This file
├── CONTRIBUTING.md
├── LICENSE                       # MIT
└── README.md
```

---

## Key Package Responsibilities

### `internal/engine`
The heart of the tool. Accepts a list of `Channel` implementations, runs them concurrently with a shared context and per-channel timeout, collects `Result` structs, and passes them to the aggregator. No knowledge of specific channels.

### `internal/channels`
Each subdirectory is a self-contained channel module. The only coupling to the rest of the system is via the `Channel` interface and `ChannelConfig` struct. New channels added here are registered in `registry.go` — nothing else needs to change.

### `internal/payload`
Generates the synthetic data to be exfiltrated. Supports random bytes, a fixed pattern, or a user-supplied file. Handles encryption and compression before handing off to a channel.

### `internal/listener`
The optional callback server. When running `dlpbuster serve`, this package starts DNS and HTTPS listeners and logs all received payloads to confirm receipt. Channels report `Passed` only when the listener confirms delivery.

### `internal/report`
Takes the aggregated results and renders them to the chosen format. The markdown and HTML renderers include: per-channel status, evidence log lines, bytes sent, duration, and a risk summary table.

### `cmd/`
Thin cobra command wrappers. Each file maps to one subcommand. All business logic lives in `internal/` — commands just parse flags, build config, and call into the engine or report package.

---

## Data Flow

```
CLI flags + config.yaml
        │
        ▼
   config.Config
        │
        ├──► payload.Generator  →  []byte (encrypted, compressed, chunked)
        │
        └──► engine.Run(channels, payload, config)
                    │
                    ├── dns.Channel.Run()    → Result
                    ├── https.Channel.Run()  → Result
                    ├── icmp.Channel.Run()   → Result
                    └── ...
                    │
                    ▼
             engine.Aggregate([]Result)
                    │
                    ▼
             report.Render(format)  →  stdout / file
```

---

## Config File Format (`~/.dlpbuster/config.yaml`)

```yaml
listener:
  dns_address: "your-vps.example.com"
  https_address: "https://your-vps.example.com:8443"

payload:
  size_bytes: 1024
  encrypt: true
  compress: false

timeout_seconds: 30

channels:
  dns:
    enabled: true
    resolver: "8.8.8.8:53"
    record_types: ["A", "TXT"]
    domain: "exfil.your-vps.example.com"
  https:
    enabled: true
    user_agent: "Mozilla/5.0"
    chunk_size: 256
  icmp:
    enabled: false   # requires root
  s3:
    enabled: true
    bucket: "your-test-bucket"
    region: "us-east-1"
  smtp:
    enabled: false
    server: "smtp.example.com:587"
    from: "test@example.com"
    to: "catch@your-vps.example.com"
  webhook:
    enabled: true
    url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

output:
  format: "human"   # human | json | markdown | html
  report_dir: "~/.dlpbuster/reports"
  log_dir: "~/.dlpbuster/logs"
```
