package policy

import (
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/repository"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/llm"
	"github.com/ilyakaznacheev/cleanenv"
	"strings"
)

// Prompt contains policy for compile prompting.
type Prompt struct {
	Tools map[string]tool `yaml:"tools"`
}

type tool struct {
	Model     string            `yaml:"model"`
	TypeName  string            `yaml:"typeName"`
	PrePrompt string            `yaml:"prePrompt"`
	Variables map[string]string `yaml:"variables"`
}

// MustNewPrompt parse config from file.
// If config invalid it panic.
// For future change in live.
func MustNewPrompt(policyCfg string) Prompt {
	var Prompt Prompt

	err := cleanenv.ReadConfig(policyCfg, &Prompt)
	if err != nil {
		panic(err)
	}

	return Prompt
}

// Get return prompt with replaces variables.
func (p Prompt) Get(tp repository.ChatType, vars map[string]string) (llm.Prompt, error) {
	t, ok := p.Tools[tp]
	if !ok {
		return p.defaultPrompt(vars), nil
	}

	for k, v := range t.Variables {
		t.PrePrompt = strings.ReplaceAll(t.PrePrompt, v, vars[k])
	}

	return llm.Prompt{
		Model: t.Model,
		Messages: []llm.Message{
			{
				Role: "user",
				Content: t.PrePrompt,
			},
		},
	}, nil
}

func (p Prompt) defaultPrompt(vars map[string]string) llm.Prompt {
	pr := p.Tools["default"]
	for k, v := range pr.Variables {
		pr.PrePrompt = strings.ReplaceAll(pr.PrePrompt, v, vars[k])
	}
	return llm.Prompt{
		Model:    pr.Model,
		Messages: []llm.Message{
			{
				"user",
				pr.PrePrompt,
			},
		},
	}
}
