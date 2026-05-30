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

// StreamGenerate streams the model output to w and returns the total tokens
// consumed (prompt + completion, incl. reasoning) for quota accounting.
// reasoning_effort is set to "minimal" because gpt-5 reasoning models would
// otherwise spend the completion-token budget on hidden reasoning, leaving
// almost nothing for the visible summary. include_usage makes the final chunk
// carry the token usage (that chunk has empty Choices).
func (c *Client) StreamGenerate(ctx context.Context, prompt string, w io.Writer) (int, error) {
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(2000),
		ReasoningEffort:     shared.ReasoningEffort("minimal"),
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	})
	defer stream.Close()

	var totalTokens int
	for stream.Next() {
		chunk := stream.Current()
		if chunk.Usage.TotalTokens > 0 {
			totalTokens = int(chunk.Usage.TotalTokens)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		if _, err := fmt.Fprint(w, delta); err != nil {
			return totalTokens, err
		}
		if f, ok := w.(interface{ Flush() }); ok {
			f.Flush()
		}
	}
	return totalTokens, stream.Err()
}
