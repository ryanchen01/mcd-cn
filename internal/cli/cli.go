package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ryanchen01/mcd-cn/internal/config"
	"github.com/ryanchen01/mcd-cn/internal/mcp"
)

const defaultServerURL = "https://mcp.mcd.cn/mcp-servers/mcd-mcp"

func Run(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		fmt.Fprintln(os.Stdout, usage())
		return nil
	}

	parsed, err := parseArgs(args)
	if err != nil {
		return fmt.Errorf("%w\n\n%s", err, usage())
	}

	token, err := config.LoadToken()
	if err != nil {
		return err
	}

	client := mcp.NewClient(defaultServerURL, token)
	result, err := client.CallTool(ctx, parsed.Tool, parsed.Params)
	if err != nil {
		return err
	}

	if len(result) == 0 || string(result) == "null" {
		return nil
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, result, "", "  "); err != nil {
		fmt.Fprintln(os.Stdout, string(result))
		return nil
	}

	fmt.Fprintln(os.Stdout, pretty.String())
	return nil
}

type parsedArgs struct {
	Tool   string
	Params map[string]string
}

func parseArgs(args []string) (parsedArgs, error) {
	tool := strings.TrimSpace(args[0])
	if tool == "" || strings.HasPrefix(tool, "-") {
		return parsedArgs{}, errors.New("missing tool name")
	}

	params := make(map[string]string)
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") || len(arg) == 2 {
			return parsedArgs{}, fmt.Errorf("unexpected argument: %s", arg)
		}

		key := strings.TrimPrefix(arg, "--")
		if key == "" {
			return parsedArgs{}, fmt.Errorf("invalid parameter: %s", arg)
		}

		if strings.Contains(key, "=") {
			parts := strings.SplitN(key, "=", 2)
			key = parts[0]
			value := parts[1]
			if value == "" {
				return parsedArgs{}, fmt.Errorf("missing value for --%s", key)
			}
			if _, exists := params[key]; exists {
				return parsedArgs{}, fmt.Errorf("duplicate parameter: --%s", key)
			}
			params[key] = value
			continue
		}

		value := "true"
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			value = args[i+1]
			i++
		}

		if _, exists := params[key]; exists {
			return parsedArgs{}, fmt.Errorf("duplicate parameter: --%s", key)
		}
		params[key] = value
	}

	return parsedArgs{Tool: tool, Params: params}, nil
}

func isHelp(arg string) bool {
	switch strings.ToLower(arg) {
	case "-h", "--help", "help":
		return true
	default:
		return false
	}
}

func usage() string {
	tools := []string{
		"campaign-calender",
		"available-coupons",
		"auto-bind-coupons",
		"my-coupons",
		"now-time-info",
	}
	sort.Strings(tools)
	return fmt.Sprintf(`Usage:
  mcd-cn <tool-name> [--param value] [--param=value] [--flag]

Examples:
  mcd-cn campaign-calender
  mcd-cn campaign-calender --specifiedDate 2025-12-09
  mcd-cn available-coupons

Notes:
  - Set MCDCN_MCP_TOKEN or provide it in .env.
  - Known tools: %s
`, strings.Join(tools, ", "))
}
