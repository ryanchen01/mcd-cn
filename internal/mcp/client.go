package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const protocolVersion = "2025-06-18"

type Client struct {
	url         string
	token       string
	httpClient  *http.Client
	sessionID   string
	initialized bool
	nextID      int
}

func NewClient(url, token string) *Client {
	return &Client{
		url:   url,
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		nextID: 1,
	}
}

func (c *Client) CallTool(ctx context.Context, name string, args map[string]string) (json.RawMessage, error) {
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, err
	}

	params := map[string]any{
		"name": name,
	}
	if len(args) > 0 {
		params["arguments"] = args
	}

	return c.doRPC(ctx, "tools/call", params, true)
}

func (c *Client) ensureInitialized(ctx context.Context) error {
	if c.initialized {
		return nil
	}

	params := map[string]any{
		"protocolVersion": protocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "mcd-cn",
			"version": "0.1.2",
		},
	}

	if _, err := c.doRPC(ctx, "initialize", params, true); err != nil {
		return err
	}

	if _, err := c.doRPC(ctx, "initialized", nil, false); err != nil {
		return err
	}

	c.initialized = true
	return nil
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (c *Client) doRPC(ctx context.Context, method string, params any, expectResponse bool) (json.RawMessage, error) {
	id := ""
	if expectResponse {
		id = strconv.Itoa(c.nextID)
		c.nextID++
	}

	request := rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	if expectResponse {
		request.ID = id
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+c.token)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json, text/event-stream")
	httpRequest.Header.Set("User-Agent", "mcd-cn/0.1.0")
	if c.sessionID != "" {
		httpRequest.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	c.captureSessionID(resp)

	if !expectResponse {
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("mcp notification failed: %s", resp.Status)
		}
		return nil, nil
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") {
		return parseSSE(resp.Body, id)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(bytes.TrimSpace(body)) == 0 {
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, formatHTTPError(resp, body)
		}
		return nil, errors.New("empty response from MCP server")
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, formatHTTPError(resp, body)
		}
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("mcp error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if !idMatches(rpcResp.ID, id) {
		return nil, errors.New("unexpected response id from MCP server")
	}

	return rpcResp.Result, nil
}

func parseSSE(reader io.Reader, id string) (json.RawMessage, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)

	var dataLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if len(dataLines) == 0 {
				continue
			}
			payload := strings.Join(dataLines, "\n")
			dataLines = dataLines[:0]

			var rpcResp rpcResponse
			if err := json.Unmarshal([]byte(payload), &rpcResp); err != nil {
				continue
			}
			if rpcResp.Error != nil {
				return nil, fmt.Errorf("mcp error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
			}
			if idMatches(rpcResp.ID, id) {
				return rpcResp.Result, nil
			}
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read sse: %w", err)
	}
	return nil, errors.New("no response received from MCP server")
}

func (c *Client) captureSessionID(resp *http.Response) {
	if c.sessionID != "" {
		return
	}
	if value := strings.TrimSpace(resp.Header.Get("Mcp-Session-Id")); value != "" {
		c.sessionID = value
	}
}

func idMatches(rawID json.RawMessage, id string) bool {
	if len(rawID) == 0 {
		return false
	}
	if rawID[0] == '"' {
		var decoded string
		if err := json.Unmarshal(rawID, &decoded); err != nil {
			return false
		}
		return decoded == id
	}
	return string(rawID) == id
}

func formatHTTPError(resp *http.Response, body []byte) error {
	snippet := compactSnippet(string(body), 300)
	if snippet == "" {
		return fmt.Errorf("mcp error: %s", resp.Status)
	}
	return fmt.Errorf("mcp error: %s: %s", resp.Status, snippet)
}

func compactSnippet(value string, max int) string {
	compact := strings.Join(strings.Fields(value), " ")
	if compact == "" {
		return ""
	}
	if len(compact) > max {
		return compact[:max] + "..."
	}
	return compact
}
