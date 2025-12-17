package entities

import (
	"net/url"
	"time"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/google/uuid"
)

type ResultMedia struct {
	ID   uuid.UUID
	Name string
	Type string
	URL  *url.URL

	// Size of file in KB
	Size      int64
	ResultID  uuid.UUID
	UserID    *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewResultMedia(
	name string,
	resultType string,
	resultURL *url.URL,
	size int64,
	resultID uuid.UUID,
	userID *uuid.UUID,
) *ResultMedia {
	return &ResultMedia{
		ID:        uuid.New(),
		Name:      name,
		Type:      resultType,
		URL:       resultURL,
		Size:      size,
		ResultID:  resultID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (r *ResultMedia) Validate() error {
	if len(r.Name) > maxLenMediaName {
		return domain.NewValidationError(nil, "name", "name is too long")
	}

	if r.Type == "" {
		return domain.NewValidationError(nil, "type", "type is required")
	}

	if r.URL.String() == "" {
		return domain.NewValidationError(nil, "url", "url is required")
	}

	if r.Size > maxSize {
		return domain.NewValidationError(nil, "size", "size is too long")
	}

	if r.ResultID == uuid.Nil {
		return domain.NewValidationError(nil, "result_id", "result_id is required")
	}

	if r.CreatedAt.After(r.UpdatedAt) {
		return domain.NewValidationError(nil, "created_at", "createdAt is in the future")
	}

	return nil
}
