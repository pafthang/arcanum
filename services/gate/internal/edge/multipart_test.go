package edge

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/pafthang/arcanum/pkg/mini"
)

func TestParseMultipartBody(t *testing.T) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("folder", "docs/2024")
	_ = w.WriteField("collection", "uploads")
	_ = w.WriteField("overwrite", "1")
	part, err := w.CreateFormFile("file", ".keep")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("keep\n")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	ct := w.FormDataContentType()

	file, form, ok, err := parseMultipartBody(buf.Bytes(), ct)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected multipart")
	}
	if string(file.Data) != "keep\n" {
		t.Fatalf("file data: %q", file.Data)
	}
	if file.Filename != ".keep" {
		t.Fatalf("filename: %q", file.Filename)
	}
	if file.Field != "file" {
		t.Fatalf("field: %q", file.Field)
	}
	if form["folder"] != "docs/2024" {
		t.Fatalf("folder: %q", form["folder"])
	}
	if form["collection"] != "uploads" {
		t.Fatalf("collection: %q", form["collection"])
	}
	if form["overwrite"] != "1" {
		t.Fatalf("overwrite: %q", form["overwrite"])
	}

	hdrs := mini.Headers{}
	applyMultipartHeaders(hdrs, file, form)
	if got := hdrs.Get(mini.HeaderStreamFilename); got != ".keep" {
		t.Fatalf("X-Mini-Filename: %q", got)
	}
	if got := hdrs.Get(mini.HeaderFormPrefix + "folder"); got != "docs/2024" {
		t.Fatalf("X-Mini-Form-folder: %q", got)
	}
	if ct := hdrs.Get("Content-Type"); ct == "" || ct == "multipart/form-data" {
		// file part may have empty CT; must not keep multipart
		if ct != "" && ct == "multipart/form-data" {
			t.Fatalf("content-type still multipart: %q", ct)
		}
	}
}

func TestParseMultipartBody_NotMultipart(t *testing.T) {
	body := []byte(`{"a":1}`)
	file, form, ok, err := parseMultipartBody(body, "application/json")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected not multipart")
	}
	if string(file.Data) != string(body) {
		t.Fatalf("body changed: %q", file.Data)
	}
	if form != nil {
		t.Fatalf("form: %v", form)
	}
}
