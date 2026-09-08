package requests_domain

import "github.com/gin-gonic/gin"

type IRequests interface {
	CreateRequest(c *gin.Context)
	CreateAssignment(c *gin.Context)
	ListRequests(c *gin.Context)
	ListRequestById(c *gin.Context)
	ListRequestStatusHistory(c *gin.Context)
	ListRequestRecommendations(c *gin.Context)
	ListAssginmentByRequest(c *gin.Context)
	SearchRequests(c *gin.Context)
	UpdateRequest(c *gin.Context)
	UpdateRequestStatus(c *gin.Context)
}
