package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// 从环境变量获取配置项，若未定义则使用默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

var (
	uploadPath    = getEnv("UPLOAD_PATH", "E:/projects/go/uploader/images")
	baseURL       = getEnv("BASE_URL", "https://dify.talkweb.com.cn/uploader/")
	maxUploadSize = 10 << 20 // 10MB
)

// 响应结构体
type response struct {
	URL string `json:"url"`
}

// 批量上传响应结构体
type batchResponse struct {
	URLs   []string          `json:"urls"`
	Errors map[string]string `json:"errors,omitempty"`
}

// 处理单个文件上传的函数
func handleSingleUpload(file io.Reader, filename string) (string, error) {
	// 使用时间戳防止重名冲突
	filename = fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(filename))
	dstPath := filepath.Join(uploadPath, filename)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		return "", fmt.Errorf("复制文件失败: %v", err)
	}

	return baseURL + filename, nil
}

// 单文件上传处理器
func uploadHandler(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded or incorrect field name 'image'"})
		return
	}

	// 创建上传目录
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Printf("Error creating upload directory: %v\n", err)
		c.JSON(500, gin.H{"error": "Unable to create upload directory"})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Error opening uploaded file"})
		return
	}
	defer src.Close()

	// 处理文件上传
	url, err := handleSingleUpload(src, file.Filename)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, response{URL: url})
}

// 批量上传处理器
func batchUploadHandler(c *gin.Context) {
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

	// 创建上传目录
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": "Unable to create upload directory"})
		return
	}

	// 准备响应数据
	response := batchResponse{
		URLs:   make([]string, 0, len(files)),
		Errors: make(map[string]string),
	}

	// 使用 WaitGroup 处理并发上传
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

			url, err := handleSingleUpload(file, fh.Filename)
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

func main() {
	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置文件上传大小限制
	r.MaxMultipartMemory = int64(maxUploadSize)

	// 注册路由
	r.POST("/upload", uploadHandler)
	r.POST("/upload/batch", batchUploadHandler)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         ":3000",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 优雅关闭的信号处理
	go func() {
		log.Println("Server is running on http://localhost:3000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
