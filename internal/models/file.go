package models

import (
	"encoding/json"
	"time"
)

type File struct {
	ID         int64           `json:"id"`
	Filepath   string          `json:"filepath"`
	Size       int64           `json:"size"`
	UploadedAt time.Time       `json:"uploaded_at"`
	Metadata   json.RawMessage `json:"metadata"`
}
