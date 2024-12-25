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
- `MEDIA_URL_PREFIX`: URL prefix for accessing uploaded media files
- `OCR_ENDPOINT`: Aliyun OCR API endpoint
- `ALIBABA_CLOUD_ACCESS_KEY_ID`: Aliyun access key ID
- `ALIBABA_CLOUD_ACCESS_KEY_SECRET`: Aliyun access key secret

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

### Docker command Deployment

```bash
# Build the image
docker build -t media-processor .

# Run the container
docker run -d \
  -p 3000:3000 \
  -e SERVER_PORT=3000 \
  -e UPLOAD_PATH=/app/images \
  -e MEDIA_URL_PREFIX=https://your-domain.com/media/ \
  -e ALIBABA_CLOUD_ACCESS_KEY_ID=your_key_id \
  -e ALIBABA_CLOUD_ACCESS_KEY_SECRET=your_key_secret \
  media-processor
```

### Docker Compose Deployment

```bash
# Create docker-compose.yml file
cat << 'EOF' > docker-compose.yml
services:
  media-processor:
    image: iamxiaowangye/media-processor:latest
    container_name: media-processor
    restart: always
    environment:
      - SERVER_PORT=3000
      - UPLOAD_PATH=/app/images
      - MEDIA_URL_PREFIX=https://your-domain.com/media/
      - OCR_ENDPOINT=ocr-api.cn-hangzhou.aliyuncs.com
      - ALIBABA_CLOUD_ACCESS_KEY_ID=your_key_id
      - ALIBABA_CLOUD_ACCESS_KEY_SECRET=your_key_secret
    volumes:
      - ./volumes/media-processor/images:/app/images
    ports:
      - "3000:3000"
EOF

# Run the container
docker-compose up -d
```

## Binary Deployment

### Binary Execution

#### Compile the binary for Linux

```powershell
# Windows PowerShell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o media-processor cmd/server/main.go

# Windows CMD
set GOOS=linux&& set GOARCH=amd64&& go build -o media-processor cmd/server/main.go
```

#### Prepare runtime environment

1. Create directories on server

```bash
# Create main directory
mkdir -p /opt/media-processor
cd /opt/media-processor

# Create configs and images directory
mkdir configs images
```

2. Copy config file to server and edit config file

```bash
cp configs/config.yaml.example configs/config.yaml
vim configs/config.yaml
```

3. Copy binary file to server

```bash
cp media-processor /opt/media-processor/media-processor
```

#### Run the binary on server

```bash
cd /opt/media-processor
./media-processor
```

### Binary as a service

```bash
# Create systemd service file
cat << 'EOF' > /etc/systemd/system/media-processor.service
[Unit]
Description=Media Processor Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/media-processor
ExecStart=/opt/media-processor/media-processor
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
systemctl daemon-reload

# Start the service
systemctl start media-processor

# Enable the service
systemctl enable media-processor
```

## API Examples

### Single File Upload

```bash
# Upload a single image file
curl -X POST http://localhost:3000/upload \
  -F "image=@/path/to/your/image.jpg"

# Response
{
  "url": "https://your-domain.com/upload/1234567890-image.jpg"
}
```

### Batch File Upload

```bash
# Upload multiple image files
curl -X POST http://localhost:3000/upload/batch \
  -F "images=@/path/to/image1.jpg" \
  -F "images=@/path/to/image2.jpg" \
  -F "images=@/path/to/image3.jpg"

# Response
{
  "urls": [
    "https://your-domain.com/upload/1234567890-image1.jpg",
    "https://your-domain.com/upload/1234567891-image2.jpg",
    "https://your-domain.com/upload/1234567892-image3.jpg"
  ]
}
```

### OCR Text Recognition

```bash
# Extract text from an image URL
curl -X POST http://localhost:3000/ocr \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-domain.com/1234567890-image.jpg"
  }'

# Response
{
  "statusCode": 200,
  "headers": {
    "content-type": "application/json;charset=utf-8",
    "x-acs-request-id": "0B2607D1-62A2-5027-9A77-D7D8E69A275D"
  },
  "body": {
    "Data": {
      "Content": "检票：B9 FQ02351 长沙南站 G1003 厂 广州南站 Changshanan Guangzhounan 2024年 07 7月 29 日 09：07 开 12车 17C号 ￥314.0元 惠 二等座"
    },
    "RequestId": "0B2607D1-62A2-5027-9A77-D7D8E69A275D"
  }
}
```

## License

MIT
