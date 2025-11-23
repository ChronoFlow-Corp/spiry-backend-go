package llm

import (
	"errors"
	"io"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/revrost/go-openrouter"
)

func read(reader *openrouter.ChatCompletionStream, bufSize int) <-chan entities.Chunk {
	input := make(chan entities.Chunk, bufSize)

	go writer(input, reader)

	return input
}

func writer(input chan<- entities.Chunk, reader *openrouter.ChatCompletionStream) {
	defer close(input)

	chunk := entities.Chunk{}

	for {
		ch, err := reader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				chunk.Err = domain.ErrStreamClosed
				input <- chunk
				return
			}
			chunk.Err = err
			input <- chunk
			return
		}

		chunk = entities.Chunk{
			OpenRouterID: ch.ID,
			Annotation:   make([]entities.UrlCitation, 0),
			Media:        make([]entities.Media, 0),
		}

		if len(ch.Choices) == 0 {
			continue
		}

		for _, c := range ch.Choices {
			for _, a := range c.Delta.Annotations {
				chunk.Annotation = append(chunk.Annotation, entities.UrlCitation{
					Title:      a.URLCitation.Title,
					EndIndex:   a.URLCitation.EndIndex,
					URL:        a.URLCitation.URL,
					Content:    a.URLCitation.Content,
					StartIndex: a.URLCitation.StartIndex,
				})
			}

			if len(c.Delta.Images) != 0 {
				chunk.Type = append(chunk.Type, entities.ChunkTypeMedia)
			}

			for _, m := range c.Delta.Images {
				chunk.Media = append(chunk.Media, entities.Media{
					Type:    string(m.Type),
					Content: m.ImageURL.URL,
				})
			}

			if c.Delta.Content != "" {
				chunk.Type = append(chunk.Type, entities.ChunkTypeText)
			}

			chunk.Content += c.Delta.Content
		}

		chunk.OpenRouterID = ch.ID

		input <- chunk
	}
}
