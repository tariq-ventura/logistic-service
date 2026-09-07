package assignments_domain

import "github.com/gin-gonic/gin"

type IAssignments interface {
	ListAssginments(c *gin.Context)
	ListAssginmentById(c *gin.Context)
	UpdateAssignmentStatus(c *gin.Context)
}
