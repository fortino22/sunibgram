package deepseek

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// Client wraps the OpenAI-compatible DeepSeek chat API. DeepSeek exposes the
// same /chat/completions contract, so the openai-go SDK is reused with a
// base-URL + key override. Used only by the Fox guide chat; thread summaries
// stay on the dedicated OpenAI client.
type Client struct {
	client openai.Client
	model  shared.ChatModel
}

// defaultBaseURL is used when DEEPSEEK_BASE_URL is unset. This client is
// DeepSeek-specific, so defaulting to the public API host is safe and avoids a
// silent misconfiguration where an empty base URL sends requests nowhere (the
// answer never streams).
const defaultBaseURL = "https://api.deepseek.com"

func NewClient() *Client {
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	c := openai.NewClient(
		option.WithAPIKey(os.Getenv("DEEPSEEK_API_KEY")),
		option.WithBaseURL(baseURL),
	)
	return &Client{
		client: c,
		model:  shared.ChatModel(os.Getenv("DEEPSEEK_MODEL")),
	}
}

// StreamGenerate streams the model output to w and returns total tokens
// consumed (for quota accounting). systemPrompt carries the wiki + persona;
// userPrompt carries the question and any prior-turn context. include_usage
// makes the final (empty-choice) chunk carry token usage.
func (c *Client) StreamGenerate(ctx context.Context, systemPrompt, userPrompt string, w io.Writer) (int, error) {
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		MaxCompletionTokens: openai.Int(1500),
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
