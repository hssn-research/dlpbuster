# DLP Bypass Tester — TODO

## Phase 1: Project Bootstrap
- [x] Initialize Go module (`go mod init dlpbuster`)
- [x] Set up project directory structure
- [x] Create `Makefile` with build, test, lint targets
- [x] Add `.goreleaser.yml` for cross-platform binary builds
- [x] Set up GitHub Actions CI pipeline (lint → test → build)
- [x] Write `LICENSE` (MIT)
- [x] Write initial `README.md` with install + quickstart

## Phase 2: Core Engine
- [x] Build channel interface (`Channel` interface with `Name()`, `Run()`, `Describe()`)
- [x] Build test runner engine (parallel channel execution, timeout control)
- [x] Build result aggregator (success/fail/partial per channel)
- [x] Add JSON + human-readable output formatters
- [x] Implement global config loader (`~/.dlpbuster/config.yaml`)
- [x] Add verbosity flags (`-v`, `-vv`, `--silent`)
- [x] Write unit tests for engine and aggregator

## Phase 3: Exfil Channel Modules
### DNS Tunnel
- [x] Build DNS query exfil (base32-encoded payload in subdomains)
- [x] Support A, TXT, MX record types
- [x] Add configurable listener/resolver target
- [x] Detect DNS filtering / NXDOMAIN-based blocking

### HTTPS Covert
- [x] Exfil over HTTPS POST to controlled endpoint
- [x] Support chunked payload (evade size-based DLP)
- [x] Add User-Agent rotation
- [x] Test against common DLP fingerprint patterns (protocol-mimicry headers + chunking implemented)

### ICMP Tunnel
- [x] Encode payload in ICMP echo request data field
- [x] Add fragmentation support
- [x] Requires raw socket (document privilege requirement)

### Cloud Storage Sync
- [x] AWS S3 PUT object exfil simulation
- [x] GCP Cloud Storage upload
- [x] Azure Blob Storage upload
- [x] Use short-lived anonymous/public bucket targets (configurable via S3Endpoint / GCSEndpoint / AzureEndpoint)

### SMTP / Email
- [x] Plaintext attachment exfil
- [x] Base64 body exfil
- [x] Subject-line steganography (low-bandwidth) — capitalisation-pattern encoding via SMTPSubjectStego

### Slack / SaaS API
- [x] Webhook POST exfil (Slack, Discord, Teams)
- [x] Configurable webhook URL target

## Phase 4: Detection Evasion Techniques
- [x] Payload encryption (AES-256-GCM before exfil)
- [x] Payload compression (gzip)
- [x] Timing jitter (randomized send intervals) — JitterMs field in ChannelConfig, applied in DNS + HTTPS channels
- [x] Payload splitting across multiple requests
- [x] Protocol mimicry headers (look like legit traffic)

## Phase 5: CLI UX
- [x] `dlpbuster run` — run all channels
- [x] `dlpbuster run --channel dns,https` — run specific channels
- [x] `dlpbuster list` — list available channels + descriptions
- [x] `dlpbuster config init` — interactive config setup wizard
- [x] `dlpbuster report` — generate markdown/HTML/JSON report from last run
- [x] Progress bar (spinner via `internal/ui/progress.go`)
- [x] Color-coded results table (pass=green, fail=red, partial=yellow)

## Phase 6: Listener / Callback Server
- [x] Build `dlpbuster serve` — lightweight callback server
- [x] DNS listener (requires port 53 or dnsmasq integration)
- [x] HTTPS listener (auto TLS via Let's Encrypt or self-signed)
- [x] Confirm receipt of each channel's payload (listener ↔ channel correlation) — X-Correlation-ID echo in HTTPS listener
- [x] Log received payloads with timestamp + source IP

## Phase 7: Reporting
- [x] Markdown report with channel results + evidence
- [x] HTML report (self-contained, no dependencies)
- [x] JSON report (for integration with other tools / pipelines)
- [x] CVSS-adjacent risk rating per blocked/passed channel — risk table in Markdown + HTML reports
- [x] Executive summary section — 2-paragraph exec summary in Markdown + HTML reports

## Phase 8: Docs & Community
- [x] Full usage documentation (docs/ folder)
- [x] Per-channel setup guides
- [x] Contribution guide (`CONTRIBUTING.md`)
- [x] Issue templates (bug, feature request, new channel)
- [ ] GitHub Discussions setup (post-push — enable in repo Settings)
- [ ] Add to awesome-security-tools list (post-publish)

## Backlog (Post-MVP)
- [ ] WebSocket exfil channel
- [ ] Bluetooth / RF (out of scope for CLI, document as limitation)
- [ ] Steganography channel (image payload)
- [ ] mTLS covert channel
- [ ] Plugin system for community channels
- [ ] Proxychains / SOCKS5 support for pivoted environments
