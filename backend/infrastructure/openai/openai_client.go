package openai

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

type Client struct {
	client openai.Client
	model  shared.ChatModel
}

func NewClient() *Client {
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("OPENAI_MODEL")
	
	c := openai.NewClient(option.WithAPIKey(apiKey))
	return &Client{
		client: c,
		model:  shared.ChatModel(model),
	}
}

func (c *Client) StreamGenerate(ctx context.Context, prompt string, w io.Writer) error {
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(1200),
	})
	defer stream.Close()

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		if _, err := fmt.Fprint(w, delta); err != nil {
			return err
		}
		if f, ok := w.(interface{ Flush() }); ok {
			f.Flush()
		}
	}
	return stream.Err()
}
