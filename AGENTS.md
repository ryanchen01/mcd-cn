# AGENTS.md

`mcd-cn` is a CLI tool that utilizes the McDonalds (China) MCP, but as a CLI tool.

## File Structure
- Entry Point: `cmd/mcd-cn/main.go`
- CLI module: `internal/cli`
- Other modules `internal/<module-name>`
- MCP specs: `docs/mcp-specs.md`
- Change log: `CHANGELOG.md`
- Docs: `docs/*`

## Environment Variables
- `MCDCN_MCP_TOKEN`
If the environment variable is set, use it. If not, load `.env` as a fallback. Throw error if not found.

## Binary
Binary should be built to `bin/mcd-cn` for MacOS and Linux, and `bin/mcd-cn.exe` for Windows.
