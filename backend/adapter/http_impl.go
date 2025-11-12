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

// CallHTTPInterfaceWithParams 使用提供的参数列表执行请求，避免查库
func CallHTTPInterfaceWithParams(ctx context.Context, iface *models.Interface,
	args map[string]any,
	params []models.InterfaceParameter, ext map[string]string) ([]byte, int, error) {

	req, err := BuildHTTPRequestWithParams(ctx, iface, args, params, ext)
	if err != nil {
		return nil, 0, err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return data, resp.StatusCode, nil
}

// BuildHTTPRequestWithParams 使用提供的参数列表构建请求，避免查库
func BuildHTTPRequestWithParams(ctx context.Context,
	iface *models.Interface,
	args map[string]any,
	params []models.InterfaceParameter, ext map[string]string) (*http.Request, error) {
	if iface == nil {
		return nil, errors.New("interface is nil")
	}
	if strings.ToLower(iface.Protocol) != "http" {
		return nil, fmt.Errorf("unsupported protocol: %s", iface.Protocol)
	}
	if iface.URL == "" {
		return nil, errors.New("interface url is empty")
	}

	method := strings.ToUpper(strings.TrimSpace(iface.Method))
	if method == "" {
		method = http.MethodGet
	}

	queryVals := url.Values{}
	bodyMap := make(map[string]any)
	headers := make(http.Header)
	pathParams := make(map[string]string)

	// 构建参数索引（按 Group 分类）
	inputParams := make(map[string]models.InterfaceParameter)
	fixedParams := make([]models.InterfaceParameter, 0)
	for _, p := range params {
		if p.Group == "input" {
			inputParams[p.Name] = p
		} else if p.Group == "fixed" {
			fixedParams = append(fixedParams, p)
		}
		// output 参数不参与请求构建
	}

	// 应用用户提供的输入参数
	for name, val := range args {
		p, ok := inputParams[name]
		if ok {
			switch strings.ToLower(p.Location) {
			case "query":
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
			case "header":
				headers.Set(name, fmt.Sprintf("%v", val))
			case "path":
				pathParams[name] = fmt.Sprintf("%v", val)
			default: // body
				bodyMap[name] = val
			}
		} else {
			// 如果参数未在 input 中定义，默认放到 body
			log.Printf("undefined parameter: %s, placing in body", name)
			bodyMap[name] = val
		}
	}

	// 应用固定参数（fixed），这些参数直接使用默认值
	for _, p := range fixedParams {
		if p.DefaultValue == nil || *p.DefaultValue == "" {
			log.Printf("Warning: fixed parameter %s has no default value", p.Name)
			continue
		}
		// 转换默认值
		convertedVal, err := convertDefaultValue(*p.DefaultValue, p.Type)
		if err != nil {
			log.Printf("Warning: failed to convert fixed parameter %s: %v", p.Name, err)
			continue
		}

		// 根据位置放置参数
		switch strings.ToLower(p.Location) {
		case "query":
			queryVals.Set(p.Name, fmt.Sprintf("%v", convertedVal))
		case "header":
			headers.Set(p.Name, fmt.Sprintf("%v", convertedVal))
		case "path":
			pathParams[p.Name] = fmt.Sprintf("%v", convertedVal)
		default: // body
			bodyMap[p.Name] = convertedVal
		}
	}

	// 如果是 GET/HEAD，将 body 参数移到 query
	if method == http.MethodGet || method == http.MethodHead {
		for k, v := range bodyMap {
			queryVals.Set(k, fmt.Sprintf("%v", v))
		}
		bodyMap = map[string]any{}
	}

	// 替换 URL 中的 path 参数
	finalURL := iface.URL
	for name, value := range pathParams {
		// 支持 {name} 和 :name 两种格式
		finalURL = strings.ReplaceAll(finalURL, "{"+name+"}", value)
		finalURL = strings.ReplaceAll(finalURL, ":"+name, value)
	}

	// Build URL with query
	u, err := url.Parse(finalURL)
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

	if iface.AuthType == "capi" {
		secretId := headers.Get("SecretId")
		secretKey := headers.Get("SecretKey")
		if _, ok := ext["SecretId"]; ok && secretId == "" {
			secretId = ext["SecretId"]
		}
		if _, ok := ext["SecretKey"]; ok && secretKey == "" {
			secretKey = ext["SecretKey"]
		}
		if headers.Get("Host") == "" {
			return nil, errors.New("missing Host in headers for capi auth")
		}
		if headers.Get("Service") == "" {
			return nil, errors.New("missing Service in headers for capi auth")
		}
		if headers.Get("Version") == "" {
			return nil, errors.New("missing Version in headers for capi auth")
		}
		if headers.Get("Action") == "" {
			return nil, errors.New("missing Action in headers for capi auth")
		}
		if headers.Get("Region") == "" {
			return nil, errors.New("missing Region in headers for capi auth")
		}
		if secretId == "" || secretKey == "" {
			return nil, errors.New("missing SecretId or SecretKey for capi auth")
		}

		tcp := TencentCloudAPIParam{
			SecretId:  secretId,
			SecretKey: secretKey,
			Host:      headers.Get("Host"),
			Service:   headers.Get("Service"),
			Version:   headers.Get("Version"),
			Action:    headers.Get("Action"),
			Region:    headers.Get("Region"),
		}
		b, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("marshal body failed: %w", err)
		}
		tcp.Payload = string(b)
		tcHeaders, err := SignatureTCHeader(tcp)
		if err != nil {
			return nil, fmt.Errorf("generate tencent cloud api signature failed: %w", err)
		}
		body := bytes.NewBuffer(b)
		req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("https://%s", tcp.Host), body)
		if err != nil {
			return nil, err
		}
		for k, vs := range tcHeaders {
			req.Header.Add(k, vs)
		}
	}

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
