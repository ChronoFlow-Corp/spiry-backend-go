package result_collector

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/google/uuid"
)

type Collector struct {
	resultID     uuid.UUID
	userID       *uuid.UUID
	chatID       uuid.UUID
	openRouterID string
	buf          strings.Builder
	medias       []*entities.ResultMedia
}

func NewCollector(userID *uuid.UUID, resultID, chatID uuid.UUID) *Collector {
	return &Collector{
		resultID: resultID,
		userID:   userID,
		chatID:   chatID,
	}
}

func (c *Collector) AddChunk(chunk entities.Chunk) {
	c.openRouterID = chunk.OpenRouterID
	for _, t := range chunk.Type {
		switch t {
		case entities.ChunkTypeText:
			c.buf.WriteString(chunk.Content)
		case entities.ChunkTypeMedia:
			u, err := url.Parse(
				fmt.Sprintf("%s/%s/%s/", c.userID.String(), c.chatID.String(), c.resultID.String()),
			)
			if err != nil {
				panic(fmt.Sprintf("failed to parse url: %v", err))
			}

			name := fmt.Sprintf("gen_%s", uuid.New().String())

			for _, m := range chunk.Media {
				c.medias = append(
					c.medias,
					entities.NewResultMedia(
						name,
						m.Type,
						u,
						computeSize(m.Content),
						c.resultID,
						c.userID,
					),
				)
			}
		default:
			continue
		}
	}
}

func (c *Collector) Finalize() (openRouterID string, answer string) {
	return c.openRouterID, c.buf.String()
}

func computeSize(s string) int64 {
	n := len(s)
	size := n * 3 / 4

	if n >= 1 && s[n-1] == '=' {
		size--
	}

	if n >= 2 && s[n-2] == '=' {
		size--
	}

	return int64(float64(size) / 1024)
}
