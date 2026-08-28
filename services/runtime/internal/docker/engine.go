package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiPrefix = "/v1.43"

// Engine starts and stops containers via the Docker Engine API.
type Engine interface {
	Start(ctx context.Context, name, image string) (string, error)
	Stop(ctx context.Context, containerID string) error
	Exec(ctx context.Context, containerID string, cmd []string) (*ExecResult, error)
}

// ExecResult is the output of a container exec.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

// Client talks HTTP to a Docker host (unix socket or TCP).
type Client struct {
	HTTP *http.Client
	Base string
}

// New builds a client. Empty host means metadata-only mode.
func New(host string) *Client {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil
	}
	transport := &http.Transport{DisableCompression: true}
	base := "http://docker"
	switch {
	case strings.HasPrefix(host, "unix://"):
		path := strings.TrimPrefix(host, "unix://")
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", path)
		}
	case strings.HasPrefix(host, "unix:"):
		path := strings.TrimPrefix(host, "unix:")
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", path)
		}
	case strings.HasPrefix(host, "/"):
		path := host
		transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", path)
		}
	case strings.HasPrefix(host, "tcp://"):
		base = "http://" + strings.TrimPrefix(host, "tcp://")
	case strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://"):
		base = strings.TrimRight(host, "/")
	default:
		base = "http://" + host
	}
	return &Client{
		HTTP: &http.Client{Timeout: 90 * time.Second, Transport: transport},
		Base: strings.TrimRight(base, "/"),
	}
}

// Start pulls the image if needed, creates and starts a long-running container.
func (c *Client) Start(ctx context.Context, name, image string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("docker not configured")
	}
	image = strings.TrimSpace(image)
	if image == "" {
		image = "debian:bookworm-slim"
	}
	name = sanitizeName(name)
	if err := c.pull(ctx, image); err != nil {
		return "", err
	}
	body, _ := json.Marshal(map[string]any{
		"Image": image,
		"Cmd":   []string{"sleep", "infinity"},
		"Labels": map[string]string{
			"arcanum.runtime": "1",
			"arcanum.name":    name,
		},
	})
	res, err := c.do(ctx, http.MethodPost, apiPrefix+"/containers/create?name="+url.QueryEscape(name), "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("docker create %d: %s", res.StatusCode, truncate(raw))
	}
	var out struct {
		ID string `json:"Id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.ID == "" {
		return "", fmt.Errorf("docker create: bad id")
	}
	startRes, err := c.do(ctx, http.MethodPost, apiPrefix+"/containers/"+url.PathEscape(out.ID)+"/start", "", nil)
	if err != nil {
		return "", err
	}
	defer startRes.Body.Close()
	if startRes.StatusCode >= 300 && startRes.StatusCode != http.StatusNotModified {
		b, _ := io.ReadAll(startRes.Body)
		return "", fmt.Errorf("docker start %d: %s", startRes.StatusCode, truncate(b))
	}
	return out.ID, nil
}

// Stop stops a container. Missing container is not an error.
func (c *Client) Stop(ctx context.Context, containerID string) error {
	if c == nil {
		return nil
	}
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil
	}
	res, err := c.do(ctx, http.MethodPost, apiPrefix+"/containers/"+url.PathEscape(containerID)+"/stop?t=10", "", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusNotModified {
		return nil
	}
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("docker stop %d: %s", res.StatusCode, truncate(b))
	}
	return nil
}

const maxExecBytes = 64 << 10

// Exec runs cmd in a running container and captures stdout/stderr.
func (c *Client) Exec(ctx context.Context, containerID string, cmd []string) (*ExecResult, error) {
	if c == nil {
		return nil, fmt.Errorf("docker not configured")
	}
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, fmt.Errorf("container id required")
	}
	if len(cmd) == 0 {
		return nil, fmt.Errorf("cmd required")
	}
	body, _ := json.Marshal(map[string]any{
		"AttachStdout": true,
		"AttachStderr": true,
		"Cmd":          cmd,
	})
	res, err := c.do(ctx, http.MethodPost, apiPrefix+"/containers/"+url.PathEscape(containerID)+"/exec", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("docker exec create %d: %s", res.StatusCode, truncate(raw))
	}
	var created struct {
		ID string `json:"Id"`
	}
	if err := json.Unmarshal(raw, &created); err != nil || created.ID == "" {
		return nil, fmt.Errorf("docker exec create: bad id")
	}
	startBody, _ := json.Marshal(map[string]any{"Detach": false, "Tty": false})
	startRes, err := c.do(ctx, http.MethodPost, apiPrefix+"/exec/"+url.PathEscape(created.ID)+"/start", "application/json", bytes.NewReader(startBody))
	if err != nil {
		return nil, err
	}
	stream, _ := io.ReadAll(io.LimitReader(startRes.Body, maxExecBytes+8))
	startRes.Body.Close()
	if startRes.StatusCode >= 300 {
		return nil, fmt.Errorf("docker exec start %d: %s", startRes.StatusCode, truncate(stream))
	}
	out, errOut := demuxDocker(stream)
	inspRes, err := c.do(ctx, http.MethodGet, apiPrefix+"/exec/"+url.PathEscape(created.ID)+"/json", "", nil)
	exit := 0
	if err == nil {
		inspRaw, _ := io.ReadAll(inspRes.Body)
		inspRes.Body.Close()
		var insp struct {
			ExitCode int `json:"ExitCode"`
		}
		if json.Unmarshal(inspRaw, &insp) == nil {
			exit = insp.ExitCode
		}
	}
	return &ExecResult{Stdout: out, Stderr: errOut, ExitCode: exit}, nil
}

func demuxDocker(raw []byte) (stdout, stderr string) {
	if len(raw) < 8 {
		return string(raw), ""
	}
	looksMux := raw[1] == 0 && raw[2] == 0 && raw[3] == 0 && (raw[0] == 1 || raw[0] == 2)
	if !looksMux {
		return string(raw), ""
	}
	var out, errB strings.Builder
	for len(raw) >= 8 {
		stream := raw[0]
		n := int(raw[4])<<24 | int(raw[5])<<16 | int(raw[6])<<8 | int(raw[7])
		raw = raw[8:]
		if n < 0 || n > len(raw) {
			break
		}
		chunk := string(raw[:n])
		raw = raw[n:]
		if stream == 2 {
			errB.WriteString(chunk)
		} else {
			out.WriteString(chunk)
		}
	}
	return out.String(), errB.String()
}

func (c *Client) pull(ctx context.Context, image string) error {
	q := url.Values{"fromImage": {image}}
	res, err := c.do(ctx, http.MethodPost, apiPrefix+"/images/create?"+q.Encode(), "", nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("docker pull %d", res.StatusCode)
	}
	return nil
}

func (c *Client) do(ctx context.Context, method, path, ctype string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.Base+path, body)
	if err != nil {
		return nil, err
	}
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	return c.HTTP.Do(req)
}

func sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "machine"
	}
	if len(out) > 60 {
		out = out[:60]
	}
	return out
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200]
	}
	return s
}
