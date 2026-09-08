package embeddings_ollama

import (
	"context"
	"strings"
)

func (c *Client) EmbedDocument(
	ctx context.Context,
	title string,
	text string,
) ([]float32, error) {
	content := strings.TrimSpace(text)

	if title = strings.TrimSpace(title); title != "" {
		content = title + "\n" + content
	}

	return c.embed(
		ctx,
		"search_document: "+content,
	)
}
