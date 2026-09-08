package embeddings_vertex

import "context"

func (c *Client) EmbedDocument(
	ctx context.Context,
	title string,
	text string,
) ([]float32, error) {
	return c.embed(
		ctx,
		text,
		"RETRIEVAL_DOCUMENT",
		title,
	)
}
