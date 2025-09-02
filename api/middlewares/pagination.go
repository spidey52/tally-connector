package middlewares

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaginationParams struct {
	Limit     int64  `form:"limit"`
	Page      int64  `form:"page"`
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

func InsertPaginationParams(c *gin.Context) {
	c.Set("request_id", uuid.New().String())

	q := c.Request.URL.Query() // get query params

	if q.Get("limit") == "" {
		q.Set("limit", "10") // set default
	}

	if q.Get("page") == "" {
		q.Set("page", "0") // set default
	}

	c.Request.URL.RawQuery = q.Encode() // apply changes

	var paginationParams PaginationParams
	err := c.ShouldBindQuery(&paginationParams)
	if err != nil {
		log.Println("Error in binding query ", err)
		c.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid query parameters",
		})
		return
	}

	c.Set("pagination_params", paginationParams)

	c.Next()
}
