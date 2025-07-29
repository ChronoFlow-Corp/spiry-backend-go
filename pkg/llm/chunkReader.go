package llm

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"sync"
)

type responseChunk struct {
	ID       string `json:"id,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	Object   string `json:"object,omitempty"`
	Created  int64  `json:"created,omitempty"`
	Choices  []struct {
		LogProbs           *string `json:"logprobs,omitempty"`
		FinishReason       string  `json:"finish_reason,omitempty"`
		NativeFinishReason string  `json:"native_finish_reason,omitempty"`
		Index              int     `json:"index,omitempty"`
		Message            Message `json:"delta,omitempty"`
	} `json:"choices,omitempty"`
	Usage           *usage `json:"usage,omitempty"`
	CompleteMessage string
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type Chunk struct {
	ID           string
	Model        string
	Created      int64
	FinishReason string
	Messages     []struct {
		FinishReason string
		Role         string
		Content      string
	}
	Usage           usage
	CompleteMessage string
	Title string
}

type controlStrings = string

const done controlStrings = "[DONE]"

type chunkReader struct {
	src      io.ReadCloser
	scanner  *bufio.Scanner
	summary  *Chunk
	bufChunk *responseChunk
	eof      bool
	mu       *sync.Mutex
	once     sync.Once
	done     chan struct{}
}

func newChunkReader(src io.ReadCloser) *chunkReader {
	return &chunkReader{
		src:      src,
		scanner:  bufio.NewScanner(src),
		summary:  new(Chunk),
		bufChunk: new(responseChunk),
		eof:      false,
		mu:       &sync.Mutex{},
		done:     make(chan struct{}),
	}
}

func (c *chunkReader) Summary() Chunk {
	return *c.summary
}

func (c *chunkReader) Close() error {
	return c.src.Close()
}

func (c *chunkReader) IsEOF() bool {
	c.mu.Lock()
	eof := c.eof
	c.mu.Unlock()

	return eof
}

func (c *chunkReader) Wait() {
	<-c.done
}

func (c *chunkReader) Chunk() (Chunk, error) {
	for {
		if !c.scanner.Scan() {
			c.mu.Lock()
			c.eof = true
			c.mu.Unlock()
			c.once.Do(func() {
				close(c.done)
			})

			return Chunk{}, io.EOF
		}

		if err := c.scanner.Err(); err != nil {
			return Chunk{}, err
		}

		if !strings.Contains(c.scanner.Text(), "data:") {
			continue
		}

		start := strings.Index(c.scanner.Text(), " {")
		if start == -1 {
			continue
		}

		if strings.HasSuffix(c.scanner.Text(), done) {
			c.mu.Lock()
			c.eof = true
			c.mu.Unlock()
			c.once.Do(func() {
				close(c.done)
			})

			return Chunk{}, io.EOF
		}

		err := json.Unmarshal([]byte(c.scanner.Text()[start:]), c.bufChunk)
		if err != nil {
			return Chunk{}, err
		}

		if c.bufChunk.Usage != nil {
			c.summary.Usage = *c.bufChunk.Usage
		}

		c.sum()
		var chunk Chunk
		chunk.Messages = make([]struct {
			FinishReason string
			Role         string
			Content      string
		}, 0)
		for _, v := range c.bufChunk.Choices {
			chunk.Messages = append(chunk.Messages, struct {
				FinishReason string
				Role         string
				Content      string
			}{
				FinishReason: v.FinishReason,
				Role:         v.Message.Role,
				Content:      v.Message.Content,
			})
		}
		return chunk, nil
	}
}

func (c *chunkReader) sum() {
	c.summary.ID = c.bufChunk.ID
	c.summary.Model = c.bufChunk.Model
	c.summary.Created = c.bufChunk.Created

	for _, v := range c.bufChunk.Choices {
		c.summary.CompleteMessage += v.Message.Content
	}
}
