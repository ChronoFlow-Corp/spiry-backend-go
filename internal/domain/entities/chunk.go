package entities

const (
	ChunkTypeText  = "text"
	ChunkTypeMedia = "media"
)

type Reader interface {
	Chunk() (Chunk, error)
	Title() string
	Model() *Model
}

type Chunk struct {
	Type         []string
	OpenRouterID string
	Content      string
	Annotation   []UrlCitation
	Media        []Media

	Err error
}

type UrlCitation struct {
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	URL        string `json:"url"`
}

type Media struct {
	Type    string
	Content string
}
