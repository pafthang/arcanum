package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// S3 is a SigV4 S3-compatible backend (AWS, MinIO, R2).
type S3 struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	PathStyle bool
	Client    *http.Client
}

// NewS3 builds a backend. Endpoint empty means AWS virtual-host.
func NewS3(endpoint, region, bucket, access, secret string, pathStyle bool) *S3 {
	if region == "" {
		region = "us-east-1"
	}
	if endpoint == "" {
		endpoint = "https://s3." + region + ".amazonaws.com"
		pathStyle = false
	}
	return &S3{
		Endpoint:  strings.TrimRight(endpoint, "/"),
		Region:    region,
		Bucket:    bucket,
		AccessKey: access,
		SecretKey: secret,
		PathStyle: pathStyle,
		Client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *S3) Name() string { return "s3" }

func (s *S3) objectURL(key string) (*url.URL, error) {
	key = strings.TrimPrefix(key, "/")
	var raw string
	if s.PathStyle {
		raw = s.Endpoint + "/" + s.Bucket + "/" + key
	} else {
		u, err := url.Parse(s.Endpoint)
		if err != nil {
			return nil, err
		}
		u.Host = s.Bucket + "." + u.Host
		u.Path = "/" + key
		return u, nil
	}
	return url.Parse(raw)
}

func (s *S3) Put(ctx context.Context, key, contentType string, data []byte) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return s.do(ctx, http.MethodPut, key, contentType, data)
}

func (s *S3) Get(ctx context.Context, key string) ([]byte, error) {
	u, err := s.objectURL(key)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	s.sign(req, sha256.Sum256(nil), time.Now().UTC())
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("s3 get %d: %s", res.StatusCode, truncate(body))
	}
	return body, nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	return s.do(ctx, http.MethodDelete, key, "", nil)
}

func (s *S3) do(ctx context.Context, method, key, contentType string, data []byte) error {
	u, err := s.objectURL(key)
	if err != nil {
		return err
	}
	var rdr io.Reader
	if data != nil {
		rdr = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rdr)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	sum := sha256.Sum256(data)
	s.sign(req, sum, time.Now().UTC())
	res, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("s3 %s %d: %s", method, res.StatusCode, truncate(b))
	}
	return nil
}

func (s *S3) PresignGet(_ context.Context, key, _, _ string, expiresSec int64) (string, error) {
	if expiresSec <= 0 {
		expiresSec = 900
	}
	u, err := s.objectURL(key)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	cred := s.AccessKey + "/" + dateStamp + "/" + s.Region + "/s3/aws4_request"
	q := u.Query()
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", cred)
	q.Set("X-Amz-Date", amzDate)
	q.Set("X-Amz-Expires", fmt.Sprintf("%d", expiresSec))
	q.Set("X-Amz-SignedHeaders", "host")
	u.RawQuery = q.Encode()

	canonical := strings.Join([]string{
		"GET",
		canonicalURI(u.Path),
		canonicalQuery(u.Query()),
		"host:" + u.Host + "\n",
		"host",
		"UNSIGNED-PAYLOAD",
	}, "\n")
	scope := dateStamp + "/" + s.Region + "/s3/aws4_request"
	sts := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(sha256Sum(canonical))
	sig := hex.EncodeToString(hmacSHA256(s.signingKey(dateStamp), sts))
	q.Set("X-Amz-Signature", sig)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
