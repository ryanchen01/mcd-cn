# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3] - 2026-01-17

### Added

- Human-readable output rendering for tool responses, including `now-time-info`.
- `--json` flag to force raw JSON output for scripts.

### Changed

- CLI output defaults to readable text when available, with a hint to use `--json` otherwise.
- Document `--json` usage and add a MCP token link in the README.

## [0.1.2] - 2026-01-17

### Changed

- Add contextual error messages for argument parsing, token loading, and tool calls.
- Include HTTP response snippets when MCP requests fail to aid troubleshooting.

## [0.1.1] - 2026-01-17

### Added

- Optional MCP URL override via `MCDCN_MCP_URL`.

## [0.1.0] - 2026-01-17

### Added

- MCP CLI client with tool invocation and JSON output formatting.
- MCP token loading from `MCDCN_MCP_TOKEN` or `.env`.
- GoReleaser configuration and GitHub Actions release workflow.
- Makefile for build/test/tidy/fmt and MIT license.

[0.1.3]: https://github.com/ryanchen01/mcd-cn/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/ryanchen01/mcd-cn/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/ryanchen01/mcd-cn/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/ryanchen01/mcd-cn/releases/tag/v0.1.0
