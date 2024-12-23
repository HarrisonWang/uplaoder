# Media Processor

A Go service for processing media files, including file upload and OCR capabilities.

## Features

- Single file upload
- Batch file upload
- OCR text recognition from image URL
- Support for future speech recognition

## Requirements

- Go 1.20 or higher
- Docker (for containerized deployment)
- Aliyun OCR API credentials

## Configuration

Environment variables:

- `SERVER_PORT`: Server port (default: "3000")
- `UPLOAD_PATH`: Path to store uploaded files
- `BASE_URL`: Base URL for accessing uploaded files
- `OCR_ENDPOINT`: Aliyun OCR API endpoint
- `ALIBABA_CLOUD_ACCESS_KEY_ID`: Aliyun access key ID
- `ALIBABA_CLOUD_ACCESS_KEY_SECRET`: Aliyun access key secret

## API Documentation

See `api/openapi.yaml` for detailed API documentation.

## Development

```bash
# Clone the repository
git clone https://github.com/harrisonwang/media-processor.git

# Install dependencies
go mod download

# Copy example config
cp configs/config.yaml.example configs/config.yaml

# Edit config file with your settings
vim configs/config.yaml

# Run the server
go run cmd/server/main.go
```

## Docker Deployment

```bash
# Build the image
docker build -t media-processor .

# Run the container
docker run -d \
  -p 3000:3000 \
  -e SERVER_PORT=3000 \
  -e UPLOAD_PATH=/app/images \
  -e BASE_URL=https://your-domain.com/upload/ \
  -e ALIBABA_CLOUD_ACCESS_KEY_ID=your_key_id \
  -e ALIBABA_CLOUD_ACCESS_KEY_SECRET=your_key_secret \
  media-processor
```

## License

MIT
