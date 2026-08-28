package models

// Blob is public metadata. Bytes live on disk, not in this DTO.
type Blob struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
	ActorID     string `json:"actorId,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// JSONUpload is an optional JSON body for POST when not multipart/raw.
type JSONUpload struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Data        []byte `json:"data"`
}
