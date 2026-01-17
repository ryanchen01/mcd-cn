package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ryanchen01/mcd-cn/internal/config"
	"github.com/ryanchen01/mcd-cn/internal/mcp"
)

const defaultServerURL = "https://mcp.mcd.cn/mcp-servers/mcd-mcp"
const serverURLEnvKey = "MCDCN_MCP_URL"

// Version is overridden at build time via -ldflags.
var Version = "dev"

func Run(ctx context.Context, args []string) error {
	if len(args) == 0 || isHelp(args[0]) {
		fmt.Fprintln(os.Stdout, usage())
		return nil
	}

	if isVersion(args[0]) {
		fmt.Fprintln(os.Stdout, Version)
		return nil
	}

	parsed, err := parseArgs(args)
	if err != nil {
		return fmt.Errorf("parse arguments: %w\n\n%s", err, usage())
	}

	token, err := config.LoadToken()
	if err != nil {
		return fmt.Errorf("load auth token: %w", err)
	}

	serverURL := resolveServerURL()
	client := mcp.NewClient(serverURL, token)
	result, err := client.CallTool(ctx, parsed.Tool, parsed.Params)
	if err != nil {
		return fmt.Errorf("call tool %q via %s: %w", parsed.Tool, serverURL, err)
	}

	if len(result) == 0 || string(result) == "null" {
		return nil
	}

	if parsed.OutputJSON {
		return printJSON(result)
	}

	if output, ok := renderHumanOutput(parsed.Tool, result); ok {
		writeOutput(output)
		return nil
	}

	writeOutput("No human-readable output. Re-run with --json for raw output.")
	return nil
}

type parsedArgs struct {
	Tool       string
	Params     map[string]string
	OutputJSON bool
}

func parseArgs(args []string) (parsedArgs, error) {
	tool := strings.TrimSpace(args[0])
	if tool == "" || strings.HasPrefix(tool, "-") {
		return parsedArgs{}, errors.New("missing tool name")
	}

	params := make(map[string]string)
	outputJSON := false
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
			if key == "json" {
				boolValue, err := strconv.ParseBool(value)
				if err != nil {
					return parsedArgs{}, fmt.Errorf("invalid value for --json: %s", value)
				}
				outputJSON = boolValue
				continue
			}
			if _, exists := params[key]; exists {
				return parsedArgs{}, fmt.Errorf("duplicate parameter: --%s", key)
			}
			params[key] = value
			continue
		}

		if key == "json" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				boolValue, err := strconv.ParseBool(args[i+1])
				if err != nil {
					return parsedArgs{}, fmt.Errorf("invalid value for --json: %s", args[i+1])
				}
				outputJSON = boolValue
				i++
				continue
			}
			outputJSON = true
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

	return parsedArgs{Tool: tool, Params: params, OutputJSON: outputJSON}, nil
}

func isHelp(arg string) bool {
	switch strings.ToLower(arg) {
	case "-h", "--help", "help":
		return true
	default:
		return false
	}
}

func isVersion(arg string) bool {
	switch strings.ToLower(arg) {
	case "-v", "--version", "version":
		return true
	default:
		return false
	}
}

func usage() string {
	type toolInfo struct {
		Name        string
		Description string
	}
	tools := []toolInfo{
		{
			Name:        "campaign-calender",
			Description: "Monthly marketing activity calendar (past/current/future).",
		},
		{
			Name:        "available-coupons",
			Description: "List coupons available to claim.",
		},
		{
			Name:        "auto-bind-coupons",
			Description: "Auto-claim all available coupons.",
		},
		{
			Name:        "my-coupons",
			Description: "List coupons already in your account.",
		},
		{
			Name:        "now-time-info",
			Description: "Fetch current server time details.",
		},
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
	toolLines := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolLines = append(toolLines, fmt.Sprintf("  - %s: %s", tool.Name, tool.Description))
	}
	return fmt.Sprintf(`Usage:
  mcd-cn <tool-name> [--param value] [--param=value] [--flag] [--json]
  mcd-cn version

Examples:
  mcd-cn campaign-calender
  mcd-cn campaign-calender --specifiedDate 2025-12-09
  mcd-cn available-coupons
  mcd-cn available-coupons --json

Notes:
  - Set MCDCN_MCP_TOKEN or provide it in .env.
  - Use --json for full JSON output (scripts).
  - Override MCP URL with MCDCN_MCP_URL.

Tools:
%s
`, strings.Join(toolLines, "\n"))
}

func resolveServerURL() string {
	if value := strings.TrimSpace(os.Getenv(serverURLEnvKey)); value != "" {
		return value
	}
	return defaultServerURL
}

type toolCallResult struct {
	Content []toolContent `json:"content"`
}

type toolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	URL      string `json:"url,omitempty"`
}

func renderHumanOutput(toolName string, result json.RawMessage) (string, bool) {
	if strings.EqualFold(toolName, "now-time-info") {
		if output, ok := renderNowTimeInfo(result); ok {
			return output, true
		}
	}

	var toolResult toolCallResult
	if err := json.Unmarshal(result, &toolResult); err == nil && len(toolResult.Content) > 0 {
		parts := make([]string, 0, len(toolResult.Content))
		for _, item := range toolResult.Content {
			switch strings.ToLower(item.Type) {
			case "text":
				if strings.TrimSpace(item.Text) != "" {
					parts = append(parts, item.Text)
				}
			case "image":
				if strings.TrimSpace(item.URL) != "" {
					parts = append(parts, fmt.Sprintf("[image] %s", item.URL))
				} else if strings.TrimSpace(item.MimeType) != "" {
					parts = append(parts, fmt.Sprintf("[image] %s", item.MimeType))
				} else if strings.TrimSpace(item.Data) != "" {
					parts = append(parts, "[image]")
				}
			default:
				if strings.TrimSpace(item.Text) != "" {
					parts = append(parts, item.Text)
				}
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n\n"), true
		}
	}

	var text string
	if err := json.Unmarshal(result, &text); err == nil && strings.TrimSpace(text) != "" {
		return text, true
	}

	var payload map[string]any
	if err := json.Unmarshal(result, &payload); err == nil {
		keys := make([]string, 0, len(payload))
		for key := range payload {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines := make([]string, 0, len(keys))
		for _, key := range keys {
			switch value := payload[key].(type) {
			case string:
				if strings.TrimSpace(value) != "" {
					lines = append(lines, fmt.Sprintf("%s: %s", key, value))
				}
			case float64, bool:
				lines = append(lines, fmt.Sprintf("%s: %v", key, value))
			}
		}
		if len(lines) > 0 {
			return strings.Join(lines, "\n"), true
		}
	}

	return "", false
}

type nowTimeInfoResponse struct {
	Success  bool            `json:"success"`
	Code     int             `json:"code"`
	Message  string          `json:"message"`
	DateTime string          `json:"datetime"`
	TraceID  string          `json:"traceId"`
	Data     nowTimeInfoData `json:"data"`
}

type nowTimeInfoData struct {
	Timestamp int64  `json:"timestamp"`
	DateTime  string `json:"datetime"`
	Formatted string `json:"formatted"`
	Date      string `json:"date"`
	Year      int    `json:"year"`
	Month     int    `json:"month"`
	Day       int    `json:"day"`
	DayOfWeek string `json:"dayOfWeek"`
	Timezone  string `json:"timezone"`
	Offset    string `json:"offset"`
	UTC       string `json:"utc"`
}

func renderNowTimeInfo(result json.RawMessage) (string, bool) {
	info, ok := parseNowTimeInfo(result)
	if !ok {
		return "", false
	}

	timeLabel := strings.TrimSpace(info.Data.Formatted)
	if timeLabel == "" {
		timeLabel = strings.TrimSpace(info.DateTime)
	}
	if timeLabel == "" {
		timeLabel = strings.TrimSpace(info.Data.DateTime)
	}

	tzLabel := strings.TrimSpace(info.Data.Timezone)
	if tzLabel == "" {
		tzLabel = strings.TrimSpace(info.Data.Offset)
	}

	lines := make([]string, 0, 5)
	if timeLabel != "" {
		if tzLabel != "" {
			lines = append(lines, fmt.Sprintf("Server time: %s (%s)", timeLabel, tzLabel))
		} else {
			lines = append(lines, fmt.Sprintf("Server time: %s", timeLabel))
		}
	}

	if info.Data.Date != "" {
		if info.Data.DayOfWeek != "" {
			lines = append(lines, fmt.Sprintf("Date: %s (%s)", info.Data.Date, info.Data.DayOfWeek))
		} else {
			lines = append(lines, fmt.Sprintf("Date: %s", info.Data.Date))
		}
	}

	if info.Data.UTC != "" {
		lines = append(lines, fmt.Sprintf("UTC: %s", info.Data.UTC))
	}

	if info.Data.Timestamp != 0 {
		lines = append(lines, fmt.Sprintf("Timestamp: %d", info.Data.Timestamp))
	}

	if info.TraceID != "" {
		lines = append(lines, fmt.Sprintf("Trace ID: %s", info.TraceID))
	}

	if len(lines) == 0 {
		return "", false
	}

	return strings.Join(lines, "\n"), true
}

func parseNowTimeInfo(result json.RawMessage) (nowTimeInfoResponse, bool) {
	var info nowTimeInfoResponse
	if err := json.Unmarshal(result, &info); err == nil && info.Data.Date != "" {
		return info, true
	}

	var toolResult toolCallResult
	if err := json.Unmarshal(result, &toolResult); err != nil {
		return nowTimeInfoResponse{}, false
	}

	for _, item := range toolResult.Content {
		text := strings.TrimSpace(item.Text)
		if text == "" {
			continue
		}
		raw, ok := extractJSONFromText(text)
		if !ok {
			continue
		}
		if err := json.Unmarshal(raw, &info); err == nil && info.Data.Date != "" {
			return info, true
		}
	}

	return nowTimeInfoResponse{}, false
}

func extractJSONFromText(value string) (json.RawMessage, bool) {
	start := 0
	for {
		index := strings.Index(value[start:], "{")
		if index == -1 {
			return nil, false
		}
		index += start
		decoder := json.NewDecoder(strings.NewReader(value[index:]))
		decoder.UseNumber()
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err == nil && len(raw) > 0 {
			return raw, true
		}
		start = index + 1
	}
}

func printJSON(result json.RawMessage) error {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, result, "", "  "); err != nil {
		writeOutput(string(result))
		return nil
	}

	writeOutput(pretty.String())
	return nil
}

func writeOutput(output string) {
	if strings.HasSuffix(output, "\n") {
		fmt.Fprint(os.Stdout, output)
		return
	}
	fmt.Fprintln(os.Stdout, output)
}
