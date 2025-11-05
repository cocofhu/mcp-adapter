package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mcp-adapter/backend/models"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPOptions 解析 Interface.Options 的结构
// 示例见用户提供的 JSON
type HTTPOptions struct {
	Method         string          `json:"method"`
	Parameters     []HTTPParam     `json:"parameters"`
	DefaultParams  []HTTPParamVal  `json:"defaultParams"`
	DefaultHeaders []HTTPHeaderVal `json:"defaultHeaders"`
}

type HTTPParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Location    string `json:"location"` // body|query|header
	Description string `json:"description"`
}

type HTTPParamVal struct {
	Name        string  `json:"name"`
	Value       any     `json:"value"`
	Type        string  `json:"type"`
	Location    string  `json:"location"` // body|query|header
	Description *string `json:"description"`
}

type HTTPHeaderVal struct {
	Name        string  `json:"name"`
	Value       string  `json:"value"`
	Description *string `json:"description"`
}

// BuildHTTPRequest 根据 Interface 和参数构建 http.Request
func BuildHTTPRequest(ctx context.Context, iface *models.Interface, args map[string]any) (*http.Request, error) {
	if iface == nil {
		return nil, errors.New("interface is nil")
	}
	if strings.ToLower(iface.Protocol) != "http" {
		return nil, fmt.Errorf("unsupported protocol: %s", iface.Protocol)
	}
	if iface.URL == "" {
		return nil, errors.New("interface url is empty")
	}

	var opts HTTPOptions
	if strings.TrimSpace(iface.Options) != "" {
		if err := json.Unmarshal([]byte(iface.Options), &opts); err != nil {
			return nil, fmt.Errorf("invalid options json: %w", err)
		}
	}

	method := strings.ToUpper(strings.TrimSpace(opts.Method))
	if method == "" {
		method = http.MethodGet
	}

	// Prepare containers
	queryVals := url.Values{}
	bodyMap := make(map[string]any)
	headers := make(http.Header)

	// Apply default headers
	for _, h := range opts.DefaultHeaders {
		headers.Set(h.Name, h.Value)
	}
	// Apply default params
	for _, p := range opts.DefaultParams {
		switch strings.ToLower(p.Location) {
		case "query":
			queryVals.Set(p.Name, fmt.Sprintf("%v", p.Value))
		case "header":
			headers.Set(p.Name, fmt.Sprintf("%v", p.Value))
		default: // body
			bodyMap[p.Name] = p.Value
		}
	}

	// Validate and apply provided args according to parameter definitions
	paramIndex := make(map[string]HTTPParam)
	for _, pd := range opts.Parameters {
		paramIndex[pd.Name] = pd
	}
	for name, val := range args {
		pd, ok := paramIndex[name]
		if ok {
			switch strings.ToLower(pd.Location) {
			case "query":
				queryVals.Set(name, fmt.Sprintf("%v", val))
			case "header":
				headers.Set(name, fmt.Sprintf("%v", val))
			default: // body
				bodyMap[name] = val
			}
		} else {
			// If not defined, default to body
			bodyMap[name] = val
		}
	}
	// Check required params
	for _, pd := range opts.Parameters {
		if pd.Required {
			_, provided := args[pd.Name]
			if !provided {
				// If default covers it, allow
				covered := false
				for _, dp := range opts.DefaultParams {
					if dp.Name == pd.Name {
						covered = true
						break
					}
				}
				if !covered {
					return nil, fmt.Errorf("missing required param: %s", pd.Name)
				}
			}
		}
	}

	// If GET/HEAD, move body params into query
	if method == http.MethodGet || method == http.MethodHead {
		for k, v := range bodyMap {
			queryVals.Set(k, fmt.Sprintf("%v", v))
		}
		bodyMap = map[string]any{}
	}

	// Build URL with query
	u, err := url.Parse(iface.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	// merge existing query
	if existing := u.Query(); len(existing) > 0 {
		for k, vs := range existing {
			for _, v := range vs {
				queryVals.Add(k, v)
			}
		}
	}
	u.RawQuery = queryVals.Encode()

	// Body
	var body io.Reader
	if len(bodyMap) > 0 {
		b, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("marshal body failed: %w", err)
		}
		body = bytes.NewBuffer(b)
		if headers.Get("Content-Type") == "" {
			headers.Set("Content-Type", "application/json")
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	// Apply headers
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	log.Printf("%+v", req)
	return req, nil
}

// CallHTTPInterface 执行请求并返回响应内容和状态码
func CallHTTPInterface(ctx context.Context, iface *models.Interface, args map[string]any) ([]byte, int, error) {
	req, err := BuildHTTPRequest(ctx, iface, args)
	if err != nil {
		return nil, 0, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}
