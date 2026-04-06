# dlpbuster

> Automated DLP control bypass testing across DNS, HTTPS, ICMP, cloud storage, email, and SaaS exfiltration channels.

[![CI](https://github.com/hssn-research/dlpbuster/actions/workflows/ci.yml/badge.svg)](https://github.com/hssn-research/dlpbuster/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hssn-research/dlpbuster)](https://goreportcard.com/report/github.com/hssn-research/dlpbuster)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## Overview

`dlpbuster` is an open-source CLI tool for red teams, penetration testers, and security engineers that **systematically verifies whether DLP controls actually prevent data leaving your environment**.

Most DLP deployments are configured — few are *continuously verified*. `dlpbuster` runs controlled, non-destructive exfiltration simulations across multiple protocols and reports exactly which paths succeeded, which were blocked, and which produced only partial detection — with CVSS-adjacent risk ratings and executive-ready output built in.

```
dlpbuster v0.1.0 — DLP Bypass Tester
[!] For authorized testing only. Obtain written permission before use.

Running 5 channels | Payload: 1024 bytes | Timeout: 30s

  CHANNEL         STATUS    DURATION   EVIDENCE
  ──────────────────────────────────────────────────────────────────────
  dns             ✓ PASSED    342ms    2/2 DNS queries passed NXDOMAIN via 8.8.8.8:53
  https           ✗ BLOCKED   1.2s     Connection reset by peer
  icmp            ~ PARTIAL   890ms    2/5 packets received
  s3              ✓ PASSED    440ms    PUT https://... → 200 OK
  webhook         ✓ PASSED    210ms    POST https://... → 200 OK

Summary: 3 passed  |  1 blocked  |  1 partial  |  0 errors  |  0 skipped
Report: ~/.dlpbuster/reports/dlpbuster-report-2026-04-06.md
```

---

## Legal Notice

This tool is for **authorized security testing only**. You are solely responsible for obtaining written permission from the system or network owner before use. Unauthorized use against systems you do not own or have explicit written permission to test is illegal and may result in criminal prosecution.

The authors accept no liability for misuse.

---

## Installation

### Binary (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/hssn-research/dlpbuster/main/scripts/install.sh | bash
```

### Go install

```bash
go install github.com/hssn-research/dlpbuster/cmd/dlpbuster@latest
```

### Build from source

```bash
git clone https://github.com/hssn-research/dlpbuster
cd dlpbuster
make build
./bin/dlpbuster --version
```

**Requirements:** Go 1.22 or later, GNU Make

---

## Quick Start

**1. Initialize configuration**

```bash
dlpbuster config init
```

Walks through setting up your callback listener address and enabling channels.

**2. Start the callback listener (on your controlled VPS)**

```bash
dlpbuster serve --dns --https
```

**3. Run all enabled channels**

```bash
dlpbuster run
```

**4. Run specific channels**

```bash
dlpbuster run --channel dns,https,s3
```

**5. Generate a report**

```bash
dlpbuster run --format html --out ./report.html
```

---

## Channels

| Channel        | Protocol    | Requires Root | Status    |
|----------------|-------------|---------------|-----------|
| DNS Tunnel     | DNS A/TXT   | No            | Available |
| HTTPS Covert   | HTTPS POST  | No            | Available |
| ICMP Tunnel    | ICMP echo   | Yes           | Available |
| AWS S3         | HTTPS       | No            | Available |
| GCP Storage    | HTTPS       | No            | Available |
| Azure Blob     | HTTPS       | No            | Available |
| SMTP / Email   | SMTP        | No            | Available |
| Webhook        | HTTPS POST  | No            | Available |
| WebSocket      | WS/WSS      | No            | Planned   |
| Steganography  | varies      | No            | Planned   |

---

## CLI Reference

```
Usage:
  dlpbuster [command] [flags]

Commands:
  run       Run exfiltration channel tests
  serve     Start the callback listener server
  list      List available channels and descriptions
  report    Render a report from the last run
  config    Manage configuration

Flags:
  --channel string     Comma-separated list of channels to run (default: all enabled)
  --payload-size int   Synthetic payload size in bytes (default: from config)
  --timeout int        Per-channel timeout in seconds (default: from config)
  --format string      Output format: human | json | markdown | html (default: human)
  --out string         Write report to file instead of stdout
  --encrypt            AES-256-GCM encrypt payload before exfiltration
  --compress           Gzip compress payload before exfiltration
  --payload-file       Use a real file as payload instead of random bytes
  -v, --verbose        Verbose output
  --silent             Suppress all output except results
  -h, --help           Help
```

---

## Configuration

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
    user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    chunk_size: 256
  icmp:
    enabled: false     # requires root / CAP_NET_RAW
    target: "your-vps.example.com"
  s3:
    enabled: true
    bucket: "your-test-bucket"
    region: "us-east-1"
    # endpoint: "http://localhost:9000"   # override for MinIO / anonymous buckets
  gcs:
    enabled: false
    bucket: "your-gcs-bucket"
  azure:
    enabled: false
    account: "youraccount"
    container: "yourcontainer"
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
```

---

## Output Formats

```bash
dlpbuster run --format json                         # machine-parseable, stdout
dlpbuster run --format markdown --out ./report.md
dlpbuster run --format html    --out ./report.html
```

HTML and Markdown reports include:

- Per-channel result table with timing and evidence
- Executive summary paragraph
- CVSS-adjacent risk rating and remediation recommendation per channel

---

## How It Works

1. Generates a **synthetic payload** of random bytes (configurable size). Never reads production data unless `--payload-file` is explicitly set.
2. Optionally encrypts (AES-256-GCM) and/or compresses (gzip) the payload.
3. Runs each enabled channel **concurrently** with per-channel timeout and optional timing jitter (anti-DLP-fingerprinting).
4. Each channel attempts to deliver the payload to your controlled callback listener.
5. HTTPS channel embeds a per-run `X-Correlation-ID` header; the listener echoes it back to confirm receipt at the transport layer.
6. DNS channel distinguishes `NXDOMAIN` responses (query left the network) from `SERVFAIL`/timeouts (likely blocked by DLP or firewall).
7. Results are aggregated and rendered to your chosen format with evidence strings for each step.

---

## Detection Evasion Features

| Feature                        | Implementation                                                  |
|--------------------------------|-----------------------------------------------------------------|
| Payload encryption             | AES-256-GCM, random key per run                                 |
| Payload compression            | gzip                                                            |
| Payload splitting              | Configurable chunk size per channel                             |
| Timing jitter                  | Random delay `[0, JitterMs]` between chunks (DNS, HTTPS)        |
| Protocol mimicry               | Realistic User-Agent, Content-Type, chunk headers               |
| SMTP steganography             | Subject-line capitalisation-pattern encoding (`SMTPSubjectStego`) |
| Anonymous bucket endpoints     | Override S3/GCS/Azure base URL for public / unauthenticated tests |

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

3. Register in `internal/registry/registry.go`
4. Add a unit test in `internal/channels/<name>/<name>_test.go`
5. Add documentation in `docs/channels/<name>.md`
6. Open a pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full contribution guide.

---

## Development

```bash
make build    # compile binary to ./bin/dlpbuster
make test     # run all tests with -race detector
make lint     # run golangci-lint
make release  # goreleaser cross-platform snapshot build
```

---

## Roadmap

- WebSocket exfiltration channel
- Steganography channel (image payload encoding)
- mTLS covert channel
- Plugin system for community-contributed channels
- SOCKS5 / proxychains integration for pivoted environments
- CI-native exit codes and structured JSON output for pipeline integration

---

## Contributing

Pull requests are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

New channel implementations are the highest-value contribution. If you have identified a DLP bypass technique not currently covered, open an issue or submit a channel module directly.

---

## License

MIT — see [LICENSE](LICENSE)


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
