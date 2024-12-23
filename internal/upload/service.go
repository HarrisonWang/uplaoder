package upload

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/harrisonwang/media-processor/configs"
)

const MaxUploadSize = 10 << 20 // 10MB

type Service struct {
	uploadPath string
	baseURL    string
}

func NewService(config *configs.Config) *Service {
	return &Service{
		uploadPath: config.Upload.Path,
		baseURL:    config.Upload.BaseURL,
	}
}

func (s *Service) Upload(file io.Reader, filename string) (string, error) {
	// 使用时间戳防止重名冲突
	filename = fmt.Sprintf("%d-%s", time.Now().UnixNano(), filepath.Base(filename))
	dstPath := filepath.Join(s.uploadPath, filename)

	if err := os.MkdirAll(s.uploadPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %v", err)
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		return "", fmt.Errorf("复制文件失败: %v", err)
	}

	return s.baseURL + filename, nil
}
