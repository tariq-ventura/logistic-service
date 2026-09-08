package requests_handlers

import (
	"context"

	requests_domain "github.com/tariq-ventura/logistic-service/internal/requests/domain"
	requests_search "github.com/tariq-ventura/logistic-service/internal/requests/search"
)

func (rh *RequestHandler) indexRequest(
	ctx context.Context,
	request *requests_domain.Request,
) {
	content := requests_search.BuildDocument(request)

	embedding, err := rh.embeddings.EmbedDocument(
		ctx,
		request.ProjectName,
		content,
	)

	if err != nil {
		// El request operacional ya fue creado. Un fallo de Vertex
		// no debe convertir la creación en error HTTP 500.
		rh.logs.LogWarning(
			"request_embedding_generation_failed",
			map[string]any{
				"requestId": request.ID,
				"error":     err.Error(),
			},
		)
		return
	}

	dbError := rh.db.SaveRequestEmbedding(
		ctx,
		request.ID,
		rh.embeddings.Provider(),
		rh.embeddings.Model(),
		content,
		embedding,
	)

	if dbError != nil {
		rh.logs.LogWarning(
			"request_embedding_save_failed",
			map[string]any{
				"requestId": request.ID,
				"error":     dbError.Error,
			},
		)
	}
}
