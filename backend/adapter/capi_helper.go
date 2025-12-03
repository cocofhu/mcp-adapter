package adapter

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func hmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

type TencentCloudAPIParam struct {
	SecretId  string
	SecretKey string
	Host      string
	Service   string
	Version   string
	Action    string
	Region    string
	Payload   string
}

func SignatureTCHeader(tcp TencentCloudAPIParam) (map[string]string, error) {

	algorithm := "TC3-HMAC-SHA256"

	timestamp := time.Now().Unix()

	// step 1: build canonical request string
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-tc-action:%s\n",
		"application/json; charset=utf-8", tcp.Host, strings.ToLower(tcp.Action))
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := sha256hex(tcp.Payload)
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)

	// step 2: build string to sign
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, tcp.Service)
	hashedCanonicalRequest := sha256hex(canonicalRequest)
	string2sign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)

	// step 3: sign string
	secretDate := hmacSha256(date, "TC3"+tcp.SecretKey)
	secretService := hmacSha256(tcp.Service, secretDate)
	secretSigning := hmacSha256("tc3_request", secretService)
	signature := hex.EncodeToString([]byte(hmacSha256(string2sign, secretSigning)))

	// step 4: build authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		tcp.SecretId,
		credentialScope,
		signedHeaders,
		signature)

	headers := map[string]string{
		"Authorization":  authorization,
		"Content-Type":   "application/json; charset=utf-8",
		"Host":           tcp.Host,
		"X-TC-Action":    tcp.Action,
		"X-TC-Timestamp": fmt.Sprintf("%d", timestamp),
		"X-TC-Version":   tcp.Version,
		"X-TC-Region":    tcp.Region,
	}

	return headers, nil
}
