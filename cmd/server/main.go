package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harrisonwang/media-processor/configs"
	"github.com/harrisonwang/media-processor/internal/ocr"
	"github.com/harrisonwang/media-processor/internal/upload"
)

func main() {
	// 加载配置
	config := configs.Load()

	// 初始化OCR服务
	ocrService, err := ocr.NewService(config)
	if err != nil {
		log.Fatalf("Failed to initialize OCR service: %v", err)
	}

	// 初始化上传服务，使用配置
	uploadService := upload.NewService(config)

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	r := gin.Default()

	// 设置文件上传大小限制
	r.MaxMultipartMemory = upload.MaxUploadSize

	// 注册路由
	r.POST("/upload", upload.SingleHandler(uploadService))
	r.POST("/upload/batch", upload.BatchHandler(uploadService))
	r.POST("/ocr", ocr.Handler(ocrService))

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         ":" + config.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 优雅关闭的信号处理
	go func() {
		log.Println("Server is starting...")
		log.Println("Available APIs:")
		log.Println("  POST /upload      - Upload single file")
		log.Println("  POST /upload/batch - Upload multiple files")
		log.Println("  POST /ocr         - OCR text recognition")
		log.Printf("Server is running on %s\n", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
