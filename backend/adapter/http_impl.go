package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mcp-adapter/backend/database"
	"mcp-adapter/backend/models"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

	// 从数据库获取接口参数
	db := database.GetDB()
	var params []models.InterfaceParameter
	db.Where("interface_id = ?", iface.ID).Find(&params)

	method := strings.ToUpper(strings.TrimSpace(iface.Method))
	if method == "" {
		method = http.MethodGet
	}

	// Prepare containers
	queryVals := url.Values{}
	bodyMap := make(map[string]any)
	headers := make(http.Header)

	// 构建参数索引
	paramIndex := make(map[string]models.InterfaceParameter)
	for _, p := range params {
		paramIndex[p.Name] = p
		
		// 应用默认值
		if p.DefaultValue != nil && args[p.Name] == nil {
			args[p.Name] = *p.DefaultValue
		}
	}

	// 应用提供的参数
	for name, val := range args {
		p, ok := paramIndex[name]
		if ok {
			switch strings.ToLower(p.Location) {
			case "query":
				queryVals.Set(name, fmt.Sprintf("%v", val))
			case "header":
				headers.Set(name, fmt.Sprintf("%v", val))
			case "path":
				// Path 参数需要在 URL 中替换占位符
				// 例如: /users/{id} -> /users/123
				// 这里暂时不处理，可以后续扩展
			default: // body
				bodyMap[name] = val
			}
		} else {
			// 如果参数未定义，默认放到 body
			bodyMap[name] = val
		}
	}

	// 检查必填参数
	for _, p := range params {
		if p.Required {
			_, provided := args[p.Name]
			if !provided && p.DefaultValue == nil {
				return nil, fmt.Errorf("missing required parameter: %s", p.Name)
			}
		}
	}

	// 如果是 GET/HEAD，将 body 参数移到 query
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
