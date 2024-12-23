package ocr

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ocr_api "github.com/alibabacloud-go/ocr-api-20210707/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/harrisonwang/media-processor/configs"
)

type Service struct {
	client *ocr_api.Client
}

func NewService(config *configs.Config) (*Service, error) {
	apiConfig := &openapi.Config{
		AccessKeyId:     tea.String(config.OCR.AlibabaCloudAccessKeyID),
		AccessKeySecret: tea.String(config.OCR.AlibabaCloudAccessKeySecret),
		Endpoint:        tea.String(config.OCR.Endpoint),
	}

	client, err := ocr_api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	return &Service{client: client}, nil
}

func (s *Service) RecognizeText(imageURL string) (*ocr_api.RecognizeAllTextResponse, error) {
	request := &ocr_api.RecognizeAllTextRequest{
		Url:  tea.String(imageURL),
		Type: tea.String("Advanced"),
	}

	return s.client.RecognizeAllText(request)
}
