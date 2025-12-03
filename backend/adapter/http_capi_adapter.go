package adapter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type HTTPCAPIAdapter struct {
}

func (h HTTPCAPIAdapter) DoRequest(ctx context.Context, req mcp.CallToolRequest, parameters Parameters, meta RequestMeta) ([]byte, error) {

	if err := checkCommonParam([]string{"Host", "Service", "Version", "Action", "Region"}, parameters.HeaderParams); err != nil {
		return nil, err
	}
	secretId := req.Header.Get("TC-API-SecretId")
	secretKey := req.Header.Get("TC-API-SecretKey")
	if _, ok := parameters.HeaderParams["SecretId"]; ok && secretId == "" {
		secretId = fmt.Sprintf("%v", parameters.HeaderParams["SecretId"])
	}
	if _, ok := parameters.HeaderParams["SecretKey"]; ok && secretKey == "" {
		secretKey = fmt.Sprintf("%v", parameters.HeaderParams["SecretKey"])
	}
	if secretId == "" || secretKey == "" {
		return nil, errors.New("missing SecretId or SecretKey for capi auth")
	}
	request, payload, err := BuildCommonHttpRequest(ctx, parameters, meta)
	if err != nil {
		return nil, err
	}
	tcp := TencentCloudAPIParam{
		SecretId:  secretId,
		SecretKey: secretKey,
		Host:      fmt.Sprintf("%v", parameters.HeaderParams["Host"]),
		Service:   fmt.Sprintf("%v", parameters.HeaderParams["Service"]),
		Version:   fmt.Sprintf("%v", parameters.HeaderParams["Version"]),
		Action:    fmt.Sprintf("%v", parameters.HeaderParams["Action"]),
		Region:    fmt.Sprintf("%v", parameters.HeaderParams["Region"]),
		Payload:   string(payload),
	}
	tcHeaders, err := SignatureTCHeader(tcp)
	request.Header = http.Header{}
	for k, v := range tcHeaders {
		request.Header.Set(k, v)
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

func checkCommonParam(names []string, params map[string]any) error {
	for _, name := range names {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("missing required parameter: %s", name)
		}
	}
	return nil
}

func (h HTTPCAPIAdapter) Compatible(meta RequestMeta) bool {
	return meta.Protocol == "http" && meta.AuthType == "capi"
}
