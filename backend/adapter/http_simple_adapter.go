package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type HTTPSimpleAdapter struct {
}

func (h HTTPSimpleAdapter) DoRequest(ctx context.Context, _ mcp.CallToolRequest, parameters Parameters, meta RequestMeta) ([]byte, error) {

	request, _, err := BuildCommonHttpRequest(ctx, parameters, meta)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed: %s", string(data))
	}
	return data, nil
}

func BuildCommonHttpRequest(ctx context.Context, parameters Parameters, meta RequestMeta) (*http.Request, []byte, error) {
	bodyMap := make(map[string]any)
	queryVals := url.Values{}
	headers := make(http.Header)
	// Step 1: 解析Body参数
	for name, val := range parameters.BodyParams {
		bodyMap[name] = val
	}
	// Step 2: 解析Query参数
	for name, val := range parameters.QueryParams {
		switch v := val.(type) {
		case []string:
			for _, item := range v {
				queryVals.Add(name, item)
			}
		case []interface{}:
			for _, item := range v {
				queryVals.Add(name, fmt.Sprintf("%v", item))
			}
		default:
			queryVals.Add(name, fmt.Sprintf("%v", val))
		}
	}
	// Step 3: 解析Header参数
	for name, val := range parameters.HeaderParams {
		headers.Set(name, fmt.Sprintf("%v", val))
	}

	// Step 4: 解析Path参数
	finalURL := meta.URL
	for name, value := range parameters.PathParams {
		// 支持 {name} 和 :name 两种格式
		finalURL = strings.ReplaceAll(finalURL, "{"+name+"}", fmt.Sprintf("%v", value))
		finalURL = strings.ReplaceAll(finalURL, ":"+name, fmt.Sprintf("%v", value))
	}

	// Step 5: 将Body参数放到Query参数中
	if meta.Method == http.MethodGet || meta.Method == http.MethodHead {
		for k, v := range bodyMap {
			queryVals.Set(k, fmt.Sprintf("%v", v))
		}
		bodyMap = map[string]any{}
	}

	// Step 6: 解析URL
	u, err := url.Parse(finalURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid url: %w", err)
	}

	// Step 7: 合并Query参数
	if existing := u.Query(); len(existing) > 0 {
		for k, vs := range existing {
			for _, v := range vs {
				queryVals.Add(k, v)
			}
		}
	}
	u.RawQuery = queryVals.Encode()
	payload := make([]byte, 0)
	var body io.Reader
	if len(bodyMap) > 0 {
		b, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal body failed: %w", err)
		}
		payload = b
		body = bytes.NewBuffer(b)
	}

	request, err := http.NewRequestWithContext(ctx, meta.Method, u.String(), body)
	if err != nil {
		return nil, nil, fmt.Errorf("create request failed: %w", err)
	}
	for k, vs := range headers {
		for _, v := range vs {
			request.Header.Add(k, v)
		}
	}
	log.Printf("request: %v", payload)
	return request, payload, nil
}

func (h HTTPSimpleAdapter) Compatible(meta RequestMeta) bool {
	return meta.Protocol == "http" && meta.AuthType == "none"
}
