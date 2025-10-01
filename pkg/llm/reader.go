package llm

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"sync"
)

const done = "[DONE]"

type Chunk struct{
	ID string
	Provider string
	Model string
	Created int64
	Content Content
}

type Content struct{
	Role string
	Content string
}

type Reader struct {
	queue []Chunk
	src io.ReadCloser
	scanner *bufio.Scanner
	mu sync.RWMutex
	next chan struct{}
	done chan struct{}
}

func (r *Reader) Read() (Chunk, bool) {
	<-r.next
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.src == nil {
		return Chunk{}, false
	}

	// TODO: fix data race
	if len(r.queue) == 0 && r.src == nil {
		return Chunk{}, false
	}

	chunk := r.queue[len(r.queue)-1]
	r.queue = r.queue[:len(r.queue)-1]

	return chunk, true
}

func (r *Reader) Done() {
	<-r.done
}

func newReader(src io.ReadCloser) *Reader {
	reader := &Reader{
		src: src,
		scanner: bufio.NewScanner(src),
		mu: sync.RWMutex{},
		queue: make([]Chunk, 0),
		next: make(chan struct{}, 1),
		done: make(chan struct{}),
	}

	go reader.read()

	return reader
}

func (r *Reader) read() {
	defer func() {
		r.mu.Lock()
		r.src = nil
		r.mu.Unlock()
		close(r.next)
		close(r.done)
	}()
	defer r.src.Close()

	for r.scanner.Scan() {
		if !strings.Contains(r.scanner.Text(), "data:") {
			continue
		}

		start := strings.Index(r.scanner.Text(), " {")
		if start == -1 {
			continue
		}

		if strings.HasSuffix(r.scanner.Text(), done) {
			return
		}

		var rawChunk responseChunk

		err := json.Unmarshal([]byte(r.scanner.Text()[start:]), &rawChunk)
		if err != nil {
			continue
		}

		if len(rawChunk.Choices) == 0 {
			continue
		}

		if rawChunk.Choices[0].Message.Content == "" {
			continue
		}

		var content Content

		for _, choice := range rawChunk.Choices {
			content.Content += choice.Message.Content
			content.Role = choice.Message.Role
		}

		chunk := Chunk{
			ID: rawChunk.ID,
			Provider: rawChunk.Provider,
			Model: rawChunk.Model,
			Created: rawChunk.Created,
			Content: content,
		}

		r.mu.Lock()
		r.queue = append(r.queue, chunk)
		r.mu.Unlock()
		r.next <- struct{}{}
	}
}

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
