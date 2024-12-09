package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	uploadPath    = getEnv("UPLOAD_PATH", "/app/images")
	baseURL       = getEnv("BASE_URL", "https://voxsay.com/upload/")
	maxUploadSize = 10 << 20 // 10MB
)

// 响应结构体
type response struct {
	URL string `json:"url"`
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)

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
