package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/application/pkg/pubSub"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/config"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/aggregates"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/entities"
	"github.com/ChronoFlow-Corp/spiry-backend-go/internal/domain/models"
	"github.com/ChronoFlow-Corp/spiry-backend-go/pkg/slctx"
	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
)

type OpenRouter struct {
	cl *openrouter.Client
}

func NewOpenRouter(cfg *config.Config) *OpenRouter {
	cl := openrouter.NewClient(cfg.LLM.Key)
	return &OpenRouter{
		cl: cl,
	}
}

func (o *OpenRouter) RecognizeTool(
	ctx context.Context,
	command models.RecognizeCommand) (models.RecognizeResult, error) {
	const op = "infrastructure.llm.RecognizeTool"

	str := strings.Builder{}

	str.WriteString(fmt.Sprintf(
		"Choose best t for this user prompt \" %s \": \n",
		command.Prompt,
	))
	str.WriteString("Give title from user prompt \n")
	str.WriteString("Tools list: \n")

	mpTools := make(map[string]*entities.Tool)

	for _, t := range command.AllowedTools {
		str.WriteString("t name: " + t.Name + "\n")
		str.WriteString("t prompt: " + t.Prompt + "\n")
		mpTools[t.Name] = t
	}
	type recognize struct {
		ChosenTool string `json:"chosen_tool"`
		ChatTitle  string `json:"chat_title"`
	}

	var rec recognize

	schema, err := jsonschema.GenerateSchemaForType(rec)
	if err != nil {
		return models.RecognizeResult{}, fmt.Errorf("%s: %w", op, err)
	}

	slctx.Logger(ctx).Debug("Recognize prompt", slog.String("prompt", str.String()))

	model := lightestModel(command.AllowedModels)

	req := openrouter.ChatCompletionRequest{
		Model: model.Name,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.UserMessage(str.String()),
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "recognizing",
				Schema: schema,
				Strict: true,
			},
		},
	}

	res, err := o.cl.CreateChatCompletion(ctx, req)
	if err != nil {
		return models.RecognizeResult{}, fmt.Errorf("%s: %w", op, err)
	}

	slctx.Logger(ctx).Debug("Response from recognize tool", slog.Any("response", res))

	for _, ch := range res.Choices {
		err = json.Unmarshal([]byte(ch.Message.Content.Text), &rec)
		if err != nil {
			return models.RecognizeResult{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	t, ok := mpTools[rec.ChosenTool]
	if !ok {
		return models.RecognizeResult{}, fmt.Errorf(
			"%s: %w",
			op,
			errors.New("cannot recognize tool"),
		)
	}

	return models.RecognizeResult{
		Model: model,
		Tool:  t,
		Title: rec.ChatTitle,
	}, nil
}

func (o *OpenRouter) GenerateStreaming(
	ctx context.Context, command models.ExecuteStreaming) (*pubSub.PubSub[entities.Chunk], error) {
	const op = "infrastructure.llm.GenerateStreaming"

	req := openrouter.ChatCompletionRequest{
		Model: command.Model.Name,
	}
	cnt := openrouter.Content{}

	enrichContext(&req, command.Context)

	if len(command.Command.Medias) != 0 {
		if command.Command.Prompt != "" {
			cnt.Multi = append(cnt.Multi, openrouter.ChatMessagePart{
				Type: openrouter.ChatMessagePartTypeText,
				Text: includeTool(command.Command.Command, command.Tool),
			})
		}

		for _, m := range command.Command.Medias {
			switch m.Type {
			case entities.MediaTypeJPEG, entities.MediaTypeGIF, entities.MediaTypePNG:
				cnt.Multi = append(cnt.Multi, openrouter.ChatMessagePart{
					Type: openrouter.ChatMessagePartTypeImageURL,
					ImageURL: &openrouter.ChatMessageImageURL{
						//TODO: get from url url base64
						URL: m.URL.String(),
					},
				})
			case entities.MediaTypePDF:
				cnt.Multi = append(cnt.Multi, openrouter.ChatMessagePart{
					Type: openrouter.ChatMessagePartTypeFile,
					File: &openrouter.FileContent{
						Filename: m.Name,
						//TODO: get from url url base64
						FileData: m.URL.String(),
					},
				})
			case entities.MediaTypeMP3, entities.MediaTypeWAV:
				cnt.Multi = append(cnt.Multi, openrouter.ChatMessagePart{
					Type: openrouter.ChatMessagePartTypeInputAudio,
					InputAudio: &openrouter.ChatMessageInputAudio{
						Data: m.URL.String(),
					},
				})
			}
		}
	} else {
		cnt.Text = includeTool(command.Command.Command, command.Tool)
	}

	req.Messages = append(req.Messages, openrouter.ChatCompletionMessage{
		Role:    openrouter.ChatMessageRoleUser,
		Content: cnt,
	})

	req.Stream = true

	stream, err := o.cl.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pubSub.NewPubSub[entities.Chunk](read(stream, 10)), nil
}

func includeTool(cm *entities.Command, tool *entities.Tool) string {
	toolPrompt := tool.Prompt
	commandPrompt := cm.Prompt

	for k, v := range cm.Settings {
		placeHolder, ok := tool.Settings[k]
		if ok {
			continue
		}

		commandPrompt = strings.ReplaceAll(cm.Prompt, placeHolder, v)
	}

	return toolPrompt + commandPrompt
}

func enrichContext(
	req *openrouter.ChatCompletionRequest,
	context []aggregates.CommandResultCouple,
) {
	msgs := make([]openrouter.ChatCompletionMessage, 0, len(context))
	for _, c := range context {
		msg := openrouter.Content{}
		if len(c.CommandMedia) != 0 {
			if c.Command.Prompt != "" {
				msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
					Type: openrouter.ChatMessagePartTypeText,
					Text: c.Command.Prompt,
				})
			}
			for _, m := range c.CommandMedia {
				switch m.Type {
				case entities.MediaTypeJPEG, entities.MediaTypeGIF, entities.MediaTypePNG:
					msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
						Type: openrouter.ChatMessagePartTypeImageURL,
						ImageURL: &openrouter.ChatMessageImageURL{
							//TODO: get from url url base64
							URL: m.URL.String(),
						},
					})
				case entities.MediaTypePDF:
					msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
						Type: openrouter.ChatMessagePartTypeFile,
						File: &openrouter.FileContent{
							Filename: m.Name,
							//TODO: get from url url base64
							FileData: m.URL.String(),
						},
					})
				case entities.MediaTypeMP3, entities.MediaTypeWAV:
					msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
						Type: openrouter.ChatMessagePartTypeInputAudio,
						InputAudio: &openrouter.ChatMessageInputAudio{
							Data: m.URL.String(),
						},
					})
				}
			}
		} else {
			msg.Text = c.Command.Prompt
		}

		msgs = append(msgs, openrouter.ChatCompletionMessage{
			Role:    openrouter.ChatMessageRoleUser,
			Content: msg,
		})

		msg = openrouter.Content{}
		if len(c.ResultMedia) != 0 {
			if c.Result.Text != "" {
				msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
					Type: openrouter.ChatMessagePartTypeText,
					Text: c.Command.Prompt,
				})
			}
			for _, m := range c.ResultMedia {
				switch m.Type {
				case entities.MediaTypeJPEG, entities.MediaTypeGIF, entities.MediaTypePNG:
					msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
						Type: openrouter.ChatMessagePartTypeImageURL,
						ImageURL: &openrouter.ChatMessageImageURL{
							//TODO: get from url url base64
							URL: m.URL.String(),
						},
					})
				case entities.MediaTypePDF:
					msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
						Type: openrouter.ChatMessagePartTypeFile,
						File: &openrouter.FileContent{
							Filename: m.Name,
							//TODO: get from url url base64
							FileData: m.URL.String(),
						},
					})
				case entities.MediaTypeMP3, entities.MediaTypeWAV:
					msg.Multi = append(msg.Multi, openrouter.ChatMessagePart{
						Type: openrouter.ChatMessagePartTypeInputAudio,
						InputAudio: &openrouter.ChatMessageInputAudio{
							Data: m.URL.String(),
						},
					})
				}
			}
		} else {
			msg.Text = c.Result.Text
		}

		msgs = append(msgs, openrouter.ChatCompletionMessage{
			Role:    openrouter.ChatMessageRoleAssistant,
			Content: msg,
		})
	}

	req.Messages = append(req.Messages, msgs...)
}
