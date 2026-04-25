# Reqlog UI

A lightweight web UI for **[reqlog](https://github.com/SagarMaheshwary/reqlog)** — search and trace logs directly from your browser.

It’s designed for small teams that want **quick visibility into logs without SSH access**.

> reqlog-ui is in an early stage. Performance, security, UX will be improved in future releases.

## Features

- Search logs from browser (same power as reqlog CLI)
- Live log streaming via SSE
- API key-based authentication
- Minimal setup (single binary)
- Built with Go — lightweight and fast

## Installation

### Go Install

```bash
go install github.com/sagarmaheshwary/reqlog-ui/cmd/reqlog-ui@latest
```

### macOS / Linux

```bash
curl -sSL https://raw.githubusercontent.com/sagarmaheshwary/reqlog-ui/master/install.sh | bash
```

- Auto-detects OS/arch
- Installs latest version
- Installs to `/usr/local/bin`

Verify:

```bash
reqlog-ui --version
```

### Windows

Download from:
[https://github.com/sagarmaheshwary/reqlog-ui/releases](https://github.com/sagarmaheshwary/reqlog-ui/releases)

Then:

- unzip
- add to `PATH`

Verify:

```bash
reqlog-ui --version
```

## Usage

### 1. Start the server

```bash
HTTP_AUTH_API_KEY=your-secret-key \
REQLOG_BINARY_PATH=/usr/local/bin/reqlog \
reqlog-ui
```

Generate a secure API key using:

```bash
openssl rand -hex 32
```

### 2. Open in browser

```
http://localhost:4000
```

### 3. Authenticate

- Enter your API key
- Start searching logs

## How it works

```text
Browser → reqlog-ui → reqlog CLI → log files/containers
```

- `reqlog-ui` does **not process logs itself**
- all filtering/searching is handled by `reqlog`
- UI simply provides a browser interface

## Configuration

| Variable                   | Description                               | Default          |
| -------------------------- | ----------------------------------------- | ---------------- |
| `HTTP_AUTH_API_KEY`        | API key required to access the UI         | **required**     |
| `HTTP_SERVER_URL`          | Address the server listens on             | `localhost:4000` |
| `REQLOG_BINARY_PATH`       | Path to reqlog binary                     | `reqlog`         |
| `REQLOG_EXECUTION_TIMEOUT` | Max time allowed for log search execution | `15m`            |
| `HTTP_STREAM_TOKEN_EXPIRY` | Expiry for SSE stream tokens              | `30s`            |
| `ENV_FILE`                 | .env path                                 | `.env`           |
| `DISABLE_PRETTY_LOGS`      | disable pretty logs and output raw JSON   | `0`              |

> See `.env.example` for all available environment variables.

## Security Notes

- API key is required for all operations
- API key is stored in browser SessionStorage after login (session-only, not persistent storage)
- SSE uses short-lived tokens for streaming due to token being passed in the querystring
- Input validation is applied before executing CLI commands

> ⚠️ This tool is intended for trusted/internal environments. It should not be exposed publicly without additional security layers (reverse proxy auth, HTTPS, etc.)

## Version Compatibility

| reqlog-ui | reqlog |
| --------- | ------ |
| v0.1.0    | v0.2.1 |

> Ensure compatible versions for correct behavior.

## Contributing

Contributions, issues, and suggestions are welcome.

## License

MIT License
