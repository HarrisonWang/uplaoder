package upload

import (
	"mime/multipart"
	"sync"

	"github.com/gin-gonic/gin"
)

type Response struct {
	URL string `json:"url"`
}

type BatchResponse struct {
	URLs   []string          `json:"urls"`
	Errors map[string]string `json:"errors,omitempty"`
}

func SingleHandler(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(400, gin.H{"error": "No file uploaded or incorrect field name 'image'"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "Error opening uploaded file"})
			return
		}
		defer src.Close()

		url, err := service.Upload(src, file.Filename)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, Response{URL: url})
	}
}

func BatchHandler(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid form data"})
			return
		}

		files := form.File["images"]
		if len(files) == 0 {
			c.JSON(400, gin.H{"error": "No files uploaded"})
			return
		}

		response := BatchResponse{
			URLs:   make([]string, 0, len(files)),
			Errors: make(map[string]string),
		}

		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, fileHeader := range files {
			wg.Add(1)
			go func(fh *multipart.FileHeader) {
				defer wg.Done()

				file, err := fh.Open()
				if err != nil {
					mu.Lock()
					response.Errors[fh.Filename] = "打开文件失败: " + err.Error()
					mu.Unlock()
					return
				}
				defer file.Close()

				url, err := service.Upload(file, fh.Filename)
				mu.Lock()
				if err != nil {
					response.Errors[fh.Filename] = err.Error()
				} else {
					response.URLs = append(response.URLs, url)
				}
				mu.Unlock()
			}(fileHeader)
		}

		wg.Wait()
		c.JSON(200, response)
	}
}
