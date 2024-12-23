package ocr

import (
	"github.com/gin-gonic/gin"
)

type Request struct {
	URL string `json:"url" binding:"required"`
}

func Handler(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request: url is required"})
			return
		}

		result, err := service.RecognizeText(req.URL)
		if err != nil {
			c.JSON(500, gin.H{"error": "OCR processing failed: " + err.Error()})
			return
		}

		c.JSON(200, result)
	}
}
