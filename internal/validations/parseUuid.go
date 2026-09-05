package validations

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseUUIDParameter(
	c *gin.Context,
	parameter string,
) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(parameter))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "El identificador no es un UUID válido",
		})

		return uuid.Nil, false
	}

	return id, true
}
