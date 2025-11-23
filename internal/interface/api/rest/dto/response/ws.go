package response

type ResponseChunk struct {
	MessageID string `json:"message_id,omitempty"`
	Content   string `json:"content,omitempty"`
	ChatID    string `json:"chat_id,omitempty"`
	Role      string `json:"role,omitempty"`
	State     string `json:"state,omitempty"`
}
