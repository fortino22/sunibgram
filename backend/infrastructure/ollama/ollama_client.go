package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client struct {
	url   string
	model string
}

func NewClient() *Client {
	return &Client{
		url:   os.Getenv("OLLAMA_URL"),
		model: os.Getenv("OLLAMA_MODEL"),
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateChunk struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (c *Client) StreamGenerate(ctx context.Context, prompt string, w io.Writer) error {
	body, _ := json.Marshal(generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk generateChunk
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			continue
		}
		fmt.Fprint(w, chunk.Response)
		if f, ok := w.(interface{ Flush() }); ok {
			f.Flush()
		}
		if chunk.Done {
			break
		}
	}
	return scanner.Err()
}
