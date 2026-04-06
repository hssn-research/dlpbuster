# dlpbuster

> Automated DLP control bypass testing across DNS, HTTPS, ICMP, cloud storage, email, and SaaS exfil channels.

[![CI](https://github.com/your-org/dlpbuster/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/dlpbuster/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-org/dlpbuster)](https://goreportcard.com/report/github.com/your-org/dlpbuster)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## What It Does

`dlpbuster` is an open-source CLI tool for red teamers and pentesters that **systematically tests whether your DLP controls actually block data leaving your environment**.

Most DLP tools are configured — few are *verified*. This tool runs controlled exfiltration simulations across multiple channels and tells you exactly which ones succeeded, which were blocked, and which were only partially detected.

```
dlpbuster v0.1.0 — DLP Bypass Tester
[!] For authorized testing only. Obtain written permission before use.

Running 5 channels against target: acme-corp.internal
Payload size: 1024 bytes  |  Timeout: 30s

  DNS Tunnel       ✓ PASSED    342ms   3/3 queries received
  HTTPS Covert     ✗ BLOCKED   1.2s    Connection reset by peer
  ICMP Tunnel      ~ PARTIAL   890ms   2/5 packets received
  S3 Upload        ✓ PASSED    440ms   200 OK
  Slack Webhook    ✓ PASSED    210ms   200 OK

Summary: 3 passed  |  1 blocked  |  1 partial
Report: ./dlpbuster-report-2025-04-04.md
```

---

## ⚠️ Legal Notice

This tool is for **authorized security testing only**. You are solely responsible for obtaining written permission from the system owner before use. Unauthorized use against systems you do not own or have explicit written permission to test is illegal.

---

## Install

### Binary (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/your-org/dlpbuster/main/scripts/install.sh | bash
```

### Go install

```bash
go install github.com/your-org/dlpbuster/cmd/dlpbuster@latest
```

### Build from source

```bash
git clone https://github.com/your-org/dlpbuster
cd dlpbuster
make build
./bin/dlpbuster --version
```

---

## Quick Start

### 1. Initialize config

```bash
dlpbuster config init
```

This walks you through setting up your callback listener address and enabling channels.

### 2. Start the callback listener (on your VPS)

```bash
dlpbuster serve --dns --https
```

### 3. Run all enabled channels

```bash
dlpbuster run
```

### 4. Run specific channels

```bash
dlpbuster run --channel dns,https,s3
```

### 5. Generate a report

```bash
dlpbuster report --format html --out ./report.html
```

---

## Channels

| Channel | Protocol | Requires Root | Status |
|---|---|---|---|
| DNS Tunnel | DNS A/TXT/MX | No | ✅ MVP |
| HTTPS Covert | HTTPS POST | No | ✅ MVP |
| ICMP Tunnel | ICMP echo | Yes | ✅ MVP |
| AWS S3 | HTTPS | No | ✅ MVP |
| GCP Storage | HTTPS | No | ✅ MVP |
| Azure Blob | HTTPS | No | ✅ MVP |
| SMTP / Email | SMTP | No | ✅ MVP |
| Slack / Webhook | HTTPS | No | ✅ MVP |
| WebSocket | WS/WSS | No | 🔜 Planned |
| Steganography | varies | No | 🔜 Planned |

---

## CLI Reference

```
dlpbuster [command] [flags]

Commands:
  run       Run exfil channel tests
  serve     Start the callback listener server
  list      List available channels
  report    Render a report from the last run
  config    Manage configuration

Flags:
  --channel string    Comma-separated list of channels to run (default: all enabled)
  --payload-size int  Synthetic payload size in bytes (default: 1024)
  --timeout int       Per-channel timeout in seconds (default: 30)
  --format string     Output format: human | json | markdown | html (default: human)
  --out string        Write report to file instead of stdout
  -v, --verbose       Verbose output
  --silent            No output except results
  -h, --help          Help
```

---

## Config File

Default location: `~/.dlpbuster/config.yaml`

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
  format: "human"
  report_dir: "~/.dlpbuster/reports"
  log_dir: "~/.dlpbuster/logs"
```

---

## Output Formats

```bash
dlpbuster run --format json    # machine-parseable, stdout
dlpbuster run --format markdown --out ./report.md
dlpbuster run --format html --out ./report.html
```

---

## How It Works

1. Generates a **synthetic payload** (random bytes of configurable size — never reads real data unless `--payload-file` is set)
2. Optionally encrypts and/or compresses the payload
3. Runs each enabled channel **concurrently** with per-channel timeout
4. Each channel attempts to deliver the payload to your callback listener
5. Listener confirms receipt → channel is marked `PASSED`
6. No receipt within timeout → `BLOCKED` or `PARTIAL`
7. Results are aggregated and rendered to your chosen format

---

## Adding a New Channel

1. Create `internal/channels/<name>/<name>.go`
2. Implement the `Channel` interface:

```go
type Channel interface {
    Name()        string
    Description() string
    Run(ctx context.Context, cfg ChannelConfig) Result
}
```

3. Register in `internal/channels/registry.go`
4. Add a unit test in `internal/channels/<name>/<name>_test.go`
5. Document in `docs/channels/<name>.md`
6. Open a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide.

---

## Development

```bash
make build    # build binary to ./bin/
make test     # run all tests with -race
make lint     # run golangci-lint
make release  # goreleaser snapshot build
```

**Requirements:** Go 1.22+, Make

---

## Roadmap

- [ ] WebSocket exfil channel
- [ ] Steganography channel (image payload)
- [ ] mTLS covert channel
- [ ] Plugin system for community channels
- [ ] SOCKS5 / proxychains integration for pivoted environments
- [ ] CI-native mode (exit codes, structured JSON output for pipelines)

---

## Contributing

PRs welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

New channels are the most valuable contribution. If you've found a DLP bypass technique that isn't covered — open an issue or submit a channel module.

---

## License

MIT — see [LICENSE](LICENSE)
