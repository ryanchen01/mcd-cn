# mcd-cn

`mcd-cn` is a small CLI client for the McDonald's China MCP server.

## Requirements

- Go 1.25+
- MCP token from the McDonald's China MCP console

## Setup

The CLI reads the MCP token from the `MCDCN_MCP_TOKEN` environment variable. If it is not set, it falls back to `.env` in the project root.

Example `.env`:

```env
MCDCN_MCP_TOKEN=your_token_here
```

## Build

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

## Usage

```sh
./bin/mcd-cn <tool-name> [--param value] [--param=value] [--flag]
```

Examples:

```sh
./bin/mcd-cn campaign-calender
./bin/mcd-cn campaign-calender --specifiedDate 2025-12-09
./bin/mcd-cn available-coupons
```

## Tools

Based on the current MCP specs, available tools include:

- `campaign-calender` (optional `--specifiedDate` in `yyyy-MM-dd`)
- `available-coupons`
- `auto-bind-coupons`
- `my-coupons`
- `now-time-info`

## Development

```sh
make fmt
make test
make tidy
```
