package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var EmptyChunkError = errors.New("empty chunk")

type Prompt struct {
	Model    string    `json:"model,omitempty"`
	Messages []Message `json:"messages,omitempty"`
	Stream   bool      `json:"stream,omitempty"`
}

type ChunkReader interface {
	Read() (Chunk, bool)
	Done()
}

type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Client struct {
	c      *http.Client
	token  string
	apiURL *url.URL
}

func NewClient(token string, apiURL *url.URL) *Client {
	return &Client{
		c:      http.DefaultClient,
		token:  token,
		apiURL: apiURL,
	}
}

func (c *Client) DoStream(ctx context.Context, pr Prompt) (ChunkReader, error) {
	const op = "pkg.llm.DoStream"

	pr.Stream = true

	raw, err := json.Marshal(&pr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL.String(), bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	r.Header.Set("Authorization", "Bearer "+c.token)
	r.Header.Set("Content-Type", "application/json")

	res, err := c.c.Do(r)
	if err != nil {
		res.Body.Close()

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return newReader(res.Body), nil
}
