# Contributing to dlpbuster

Thank you for your interest in contributing to dlpbuster!

## Getting Started

```bash
git clone https://github.com/hssn-research/dlpbuster
cd dlpbuster
make build
make test
```

**Requirements:** Go 1.22+, Make

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

3. Register in `internal/channels/registry.go` (`All()` and `Lookup()`)
4. Add unit test in `internal/channels/<name>/<name>_test.go` — mock all network calls
5. Document in `docs/channels/<name>.md`
6. Open a PR with: description of the channel, bypass technique, evidence it works

## Code Standards

- Go 1.22+, `gofmt`/`goimports` formatted
- All exported types must have godoc comments
- No global mutable state
- Tests must run offline (use `httptest`, mock resolvers)
- `t.Parallel()` in all tests
- >80% coverage on core packages

## Commit Messages

```
feat(dns): add TXT record exfil mode
fix(engine): handle nil result from timed-out channel
docs(icmp): document raw socket privilege requirement
test(https): add chunked payload unit test
```

## Security & Ethics

All contributions must:
- Only generate synthetic payloads (never exfiltrate real data)
- Respect the scope boundary: DLP bypass testing only, no C2 or implants
- Pass `golangci-lint` with no new security issues

## Pull Request Process

1. Fork the repo and create a feature branch
2. Make your changes with tests
3. Run `make test lint`
4. Open a PR with a clear description

## Reporting Security Issues

Please report security vulnerabilities privately — do not open a public issue.
Email: security@example.com (update with real contact)
