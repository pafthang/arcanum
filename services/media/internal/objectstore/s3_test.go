package objectstore

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestS3PutGetDeletePresign(t *testing.T) {
	objects := map[string][]byte{}
	var lastAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastAuth = r.Header.Get("Authorization")
		key := strings.TrimPrefix(r.URL.Path, "/bucket/")
		switch r.Method {
		case http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			objects[key] = b
			w.WriteHeader(200)
		case http.MethodGet:
			if strings.Contains(r.URL.RawQuery, "X-Amz-Signature=") {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("presigned"))
				return
			}
			b, ok := objects[key]
			if !ok {
				w.WriteHeader(404)
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write(b)
		case http.MethodDelete:
			delete(objects, key)
			w.WriteHeader(204)
		default:
			w.WriteHeader(405)
		}
	}))
	defer srv.Close()

	s := NewS3(srv.URL, "us-east-1", "bucket", "AKIA", "secret", true)
	s.Client = srv.Client()
	ctx := context.Background()
	if err := s.Put(ctx, "sp1/b1", "text/plain", []byte("hello")); err != nil {
		t.Fatalf("put: %v", err)
	}
	if !strings.Contains(lastAuth, "AWS4-HMAC-SHA256") {
		t.Fatalf("missing sigv4 auth: %s", lastAuth)
	}
	got, err := s.Get(ctx, "sp1/b1")
	if err != nil || string(got) != "hello" {
		t.Fatalf("get: %v %q", err, got)
	}
	url, err := s.PresignGet(ctx, "sp1/b1", "f.txt", "text/plain", 60)
	if err != nil || !strings.Contains(url, "X-Amz-Signature=") {
		t.Fatalf("presign: %v %s", err, url)
	}
	if err := s.Delete(ctx, "sp1/b1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
