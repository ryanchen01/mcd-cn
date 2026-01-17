# mcd-cn

`mcd-cn` is a small CLI client for the McDonald's China MCP server.

## Setup

The CLI reads the MCP token from the `MCDCN_MCP_TOKEN` environment variable. If it is not set, it falls back to `.env` in the project root.

Example `.env`:

```env
MCDCN_MCP_TOKEN=your_token_here
```

[Get MCP Token](https://www.google.com/search?q=%231-%E7%94%B3%E8%AF%B7mcp-token)

## Installation

Homebrew (macOS/Linux):

```sh
brew install ryanchen01/tap/mcd-cn
```

## Usage

```sh
./bin/mcd-cn <tool-name> [--param value] [--param=value] [--flag] [--json]
```

Examples:

```sh
./bin/mcd-cn campaign-calender
./bin/mcd-cn campaign-calender --specifiedDate 2025-12-09
./bin/mcd-cn available-coupons
./bin/mcd-cn available-coupons --json
```

## Tools

Based on the current MCP specs, available tools include:

- `campaign-calender`: monthly marketing activity calendar (optional `--specifiedDate` in `yyyy-MM-dd`).
- `available-coupons`: list coupons available to claim.
- `auto-bind-coupons`: auto-claim all available coupons.
- `my-coupons`: list coupons already in your account.
- `now-time-info`: fetch current server time details.

## Development

```sh
make fmt
make test
make tidy
```

### Requirements

- Go 1.25+
- MCP token from the McDonald's China MCP console

### Build

```sh
make build
```

This produces:

- `bin/mcd-cn` on macOS/Linux
- `bin/mcd-cn.exe` on Windows

You can also build directly:

```sh
go build -o bin/mcd-cn ./cmd/mcd-cn
```