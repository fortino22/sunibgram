package openai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStreamGenerateLive(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := NewClient()
	var buf strings.Builder
	if err := c.StreamGenerate(ctx, "Reply with exactly: PONG", &buf); err != nil {
		t.Fatalf("stream: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatal("empty response")
	}
	t.Logf("model=%s response=%q", c.model, out)
}
