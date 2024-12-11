package main

import (
	"context"
	"encoding/json"
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
func handleSingleUpload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 使用时间戳防止重名冲突
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(header.Filename))
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许 POST 方法
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制请求体大小，防止过大的文件上传占用内存
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadSize))

	// 解析multipart/form-data表单
	if err := r.ParseMultipartForm(int64(maxUploadSize)); err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No file uploaded or incorrect field name 'image'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建上传目录（如果不存在则自动创建）
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Printf("Error creating upload directory: %v\n", err)
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return
	}

	// 使用时间戳防止重名冲突，并对文件名进行基本清理
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(fileHeader.Filename))
	dstPath := filepath.Join(uploadPath, filename)

	dstFile, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Error creating file: %v\n", err)
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dstFile.Close()

	// 将上传的内容写入目标文件中
	if _, err := io.Copy(dstFile, file); err != nil {
		log.Printf("Error while copying file data: %v\n", err)
		http.Error(w, "Error while copying file data", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应，包括文件的访问URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response{URL: baseURL + filename})
}

// 批量上传处理器
func batchUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制请求体大小
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxUploadSize*10)) // 允许更大的总体积
	if err := r.ParseMultipartForm(int64(maxUploadSize * 10)); err != nil {
		http.Error(w, "文件太大或表单数据无效", http.StatusBadRequest)
		return
	}

	// 确保上传目录存在
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		http.Error(w, "无法创建上传目录", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["images"] // 注意字段名为 images
	if len(files) == 0 {
		http.Error(w, "没有上传文件", http.StatusBadRequest)
		return
	}

	// 准备响应数据
	response := batchResponse{
		URLs:   make([]string, 0, len(files)),
		Errors: make(map[string]string),
	}

	// 使用 WaitGroup 处理并发上传
	var wg sync.WaitGroup
	var mu sync.Mutex // 保护响应数据的并发访问

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

			url, err := handleSingleUpload(file, fh)
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

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/upload/batch", batchUploadHandler) // 添加新的路由

	srv := &http.Server{
		Addr:         ":3000",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 优雅关闭的信号处理
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server is running on http://localhost:3000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// 等待中断信号（Ctrl+C或系统SIGTERM）
	<-stopChan
	log.Println("Shutting down server...")

	// 创建上下文，给予一定时间让当前请求处理完毕
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully.")
}
