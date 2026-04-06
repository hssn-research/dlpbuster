# HTTPS Covert Channel

## Overview
Exfiltrates payload as chunked HTTPS POST requests to a controlled endpoint.

## How It Works
1. Payload is split into configurable chunks (default 256 bytes)
2. Each chunk is POSTed to the listener with chunk index/total headers
3. TLS used for transport

## Why It Bypasses DLP
- Payload is split to avoid size-based DLP triggers
- Standard HTTPS port (443) is almost always allowed
- Mimics legitimate API traffic with configurable User-Agent

## Configuration
```yaml
channels:
  https:
    enabled: true
    user_agent: "Mozilla/5.0"
    chunk_size: 256
```

## Requirements
- No root required
- Listener: `dlpbuster serve --https`
