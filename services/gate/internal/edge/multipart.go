package edge

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
)

// multipartFile is the first file part extracted from a multipart body.
type multipartFile struct {
	Data        []byte
	Filename    string
	ContentType string
	Field       string
}

// parseMultipartBody extracts text form fields and the first file part.
// When contentType is not multipart, returns ok=false with the original body unchanged.
func parseMultipartBody(body []byte, contentType string) (file multipartFile, form map[string]string, ok bool, err error) {
	mediaType, params, perr := mime.ParseMediaType(contentType)
	if perr != nil || !strings.HasPrefix(strings.ToLower(mediaType), "multipart/") {
		return multipartFile{Data: body}, nil, false, nil
	}
	boundary := params["boundary"]
	if boundary == "" {
		return multipartFile{}, nil, false, fmt.Errorf("multipart: missing boundary")
	}
	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	form = make(map[string]string)
	var gotFile bool
	for {
		part, nerr := mr.NextPart()
		if nerr == io.EOF {
			break
		}
		if nerr != nil {
			return multipartFile{}, nil, false, fmt.Errorf("multipart: %w", nerr)
		}
		name := part.FormName()
		filename := part.FileName()
		ct := part.Header.Get("Content-Type")
		data, rerr := io.ReadAll(part)
		_ = part.Close()
		if rerr != nil {
			return multipartFile{}, nil, false, fmt.Errorf("multipart part: %w", rerr)
		}
		if filename != "" {
			if !gotFile {
				file = multipartFile{
					Data:        data,
					Filename:    filename,
					ContentType: ct,
					Field:       name,
				}
				gotFile = true
			}
			continue
		}
		if name != "" {
			form[name] = string(data)
		}
	}
	if !gotFile {
		file.Data = []byte{}
	}
	return file, form, true, nil
}

// applyMultipartHeaders injects gate multipart headers onto the NATS request.
func applyMultipartHeaders(h mini.Headers, file multipartFile, form map[string]string) {
	nh := nats.Header(h)
	// Body is the file bytes — Content-Type must describe the file, not multipart.
	if file.ContentType != "" {
		nh.Set("Content-Type", file.ContentType)
	} else {
		nh.Del("Content-Type")
	}
	if file.Filename != "" {
		nh.Set(mini.HeaderStreamFilename, file.Filename)
	}
	if file.Field != "" {
		nh.Set(mini.HeaderFileField, file.Field)
	}
	for k, v := range form {
		if k == "" {
			continue
		}
		nh.Set(mini.HeaderFormPrefix+k, v)
	}
}
