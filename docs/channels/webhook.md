# Webhook Channel

## Overview
Exfiltrates payload as a base64-encoded JSON POST to a Slack, Discord, Teams, or custom webhook URL.

## Configuration
```yaml
channels:
  webhook:
    enabled: true
    url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

## Why It Bypasses DLP
SaaS webhook destinations are typically whitelisted; JSON payloads blend with legitimate application traffic.

## Requirements
- No root required
- Outbound HTTPS to the webhook domain must be permitted
