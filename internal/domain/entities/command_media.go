package entities

import (
	"net/url"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

const (
	maxLenMediaName = 128
	maxSize         = 30000
)

const (
	MediaTypeJPEG = "image/jpeg"
	MediaTypePNG  = "image/png"
	MediaTypeGIF  = "image/gif"
	MediaTypeWEBP = "image/webp"
	MediaTypeMP4  = "video/mp4"
	MediaTypeMP3  = "audio/mpeg"
	MediaTypeWAV  = "audio/wav"
	MediaTypePDF  = "application/pdf"
)

// CommandMedia is files with command to llm.
type CommandMedia struct {
	ID   uuid.UUID
	Name string
	Type string
	URL  url.URL

	// Size of file in KB
	Size      uint64
	CommandID *uuid.UUID
	UserID    *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCommandMedia(
	name string,
	mediaType string,
	mediaURL url.URL,
	size uint64,
	commandID *uuid.UUID,
	userID *uuid.UUID,
) *CommandMedia {
	return &CommandMedia{
		ID:        uuid.New(),
		Name:      name,
		Type:      mediaType,
		URL:       mediaURL,
		Size:      size,
		CommandID: commandID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *CommandMedia) Validate() error {
	if c.Name == "" {
		return domain.NewValidationError(nil, "name", "name is required")
	}

	if c.Type == "" {
		return domain.NewValidationError(nil, "type", "type is required")
	}

	switch c.Type {
	case MediaTypeJPEG, MediaTypePNG, MediaTypeGIF, MediaTypeWEBP,
		MediaTypeMP4, MediaTypeMP3, MediaTypeWAV, MediaTypePDF:
	default:
		return domain.NewValidationError(nil, "type", "type is invalid")
	}

	if len(c.Name) > maxLenMediaName {
		return domain.NewValidationError(nil, "name", "name is too long")
	}

	if c.URL.String() == "" {
		return domain.NewValidationError(nil, "url", "url is required")
	}

	if c.Size > maxSize {
		return domain.NewValidationError(nil, "size", "size is too large")
	}

	if c.CreatedAt.After(c.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
