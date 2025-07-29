package llm

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestNewChunkReader(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty",
			data: []byte{},
		},
		{
			name: "done msg",
			data: []byte("data: [DONE]"),
		},
		{
			name: "chunk parse",
			data: []byte(`data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null,"native_finish_reason":null,"logprobs":null}],"usage":{"prompt_tokens":183,"completion_tokens":25,"total_tokens":208}}`),
		},
		{
			name: "multiply chunk parse",
			data: []byte(`data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" I"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" assist"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" you"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" today"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":"?"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}`),
		},
		{
			name: "summary msg",
			data: []byte(`data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" I"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" assist"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" you"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":" today"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}

data: {"id":"gen-1753280636-RRwXLUetXTvT0VRMWWLP","provider":"Nebius","model":"mistralai/mistral-small-3.1-24b-instruct","object":"chat.completion.chunk","created":1753280636,"choices":[{"index":0,"delta":{"role":"assistant","content":"?"},"finish_reason":null,"native_finish_reason":null,"logprobs":null}]}`),
		},
	}
	type mockCloser struct{
		io.Reader
		io.Closer
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(fmt.Sprintf("test %s", tt.name))
			m := &mockCloser{Reader: bytes.NewReader(tt.data)}
			c := newChunkReader(m)
			for {
				chunk, err := c.Chunk()
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Fatal(err)
				}
				sum := c.Summary()
				t.Log(sum)
				t.Log(chunk)
			}
		})

	}
}
