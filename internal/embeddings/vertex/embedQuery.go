package embeddings_vertex

import "context"

func (c *Client) EmbedQuery(
	ctx context.Context,
	text string,
) ([]float32, error) {
	return c.embed(
		ctx,
		text,
		"RETRIEVAL_QUERY",
		"",
	)
}
