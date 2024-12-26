#!/bin/bash

# 接收参数
DEPLOY_PATH="$1"
ACTION="$2"  # 可以是 backup, deploy

case $ACTION in
  "backup")
    # 创建备份
    mkdir -p "${DEPLOY_PATH}/backups"
    if [ -f "${DEPLOY_PATH}/media-processor" ]; then
      cp "${DEPLOY_PATH}/media-processor" "${DEPLOY_PATH}/backups/media-processor.$(date +%Y%m%d_%H%M%S)"
      if [ -f "${DEPLOY_PATH}/configs/config.yaml" ]; then
        cp "${DEPLOY_PATH}/configs/config.yaml" "${DEPLOY_PATH}/backups/config.yaml.$(date +%Y%m%d_%H%M%S)"
      fi
    else
      echo 'No existing files to backup'
    fi
    ;;

  "deploy")
    # 部署新文件
    mkdir -p "${DEPLOY_PATH}/configs"
    chmod +x "${DEPLOY_PATH}/media-processor"
    ;;

  *)
    echo "Unknown action: $ACTION"
    exit 1
    ;;
esac 