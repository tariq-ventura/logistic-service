package embeddings_ollama

import (
	"context"
	"strings"
)

func (c *Client) EmbedQuery(
	ctx context.Context,
	text string,
) ([]float32, error) {
	return c.embed(
		ctx,
		"search_query: "+strings.TrimSpace(text),
	)
}
