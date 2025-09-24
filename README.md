# File Transfer API

一个基于Go+Gin的文件转存API服务，支持将网络上的任意文件下载并上传到阿里云OSS，无文件类型和大小限制。

## 项目架构

本项目采用类似于api-server的分层架构：

```
transfer/
├── cmd/                    # 应用入口
│   └── main.go
├── app/                    # 应用层
│   ├── controller/         # 控制器
│   ├── middleware/         # 中间件
│   └── router/            # 路由
├── internal/              # 内部包
│   ├── dto/               # 数据传输对象
│   └── service/           # 业务逻辑
├── pkg/                   # 公共包
│   ├── config/            # 配置管理
│   ├── logger/            # 日志管理
│   └── resp/              # 响应工具
├── config/                # 配置文件
├── docs/                  # API文档
├── scripts/               # 脚本文件
├── go.mod
├── Makefile
└── README.md
```

## 功能特性

- 🚀 高性能的Go+Gin框架
- 📁 支持图片、视频、音频等多种媒体格式
- ☁️ 自动上传到阿里云OSS
- 🔄 内置重试机制和容错处理
- 📏 文件大小限制保护
- 🌐 CORS跨域支持
- 📊 健康检查接口
- 📖 Swagger API文档
- 📝 结构化日志记录

## 支持的媒体类型

### 图片格式
- JPEG (.jpg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)

### 视频格式
- MP4 (.mp4)
- AVI (.avi)
- MOV (.mov)

### 音频格式
- MP3 (.mp3)
- WAV (.wav)
- AAC (.aac)

## 快速开始

### 1. 环境配置

复制配置模板文件并编辑：

```bash
cp config/local.yaml.example config/local.yaml
```

然后编辑 `config/local.yaml` 文件，填入你的阿里云OSS配置：

```yaml
aliyun_oss:
  access_key_id: "你的AccessKey_ID"
  access_key_secret: "你的AccessKey_Secret"
  endpoint: "你的OSS端点"
  bucket: "你的存储桶名称"
  region: "你的区域"
```

**注意**: `config/local.yaml` 文件已添加到 `.gitignore`，不会被提交到git仓库，可以安全地存储真实的配置信息。

### 2. 安装依赖

```bash
make deps
```

### 3. 运行服务

```bash
make run
# 或者
go run cmd/main.go
```

服务将在 `http://localhost:8199` 启动。

### 4. 生成API文档

```bash
make docs
```

访问 `http://localhost:8199/swagger/index.html` 查看API文档。

## API 接口

### 健康检查
```
GET /api/v1/media/health
```

### 媒体转存
```
POST /api/v1/media/transfer
```

请求体：
```json
{
  "url": "https://example.com/image.jpg"
}
```

成功响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "message": "Media transferred successfully",
    "oss_url": "https://wavespeed-oss-xxx.oss-ap-southeast-1.oss-accesspoint.aliyuncs.com/media/2025/09/23/14-30-45/image.jpg",
    "original_url": "https://example.com/image.jpg",
    "file_size": 1024000,
    "content_type": "image/jpeg"
  }
}
```

## 使用示例

### cURL
```bash
curl -X POST http://localhost:8199/api/v1/media/transfer \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/sample.jpg"}'
```

### 测试脚本
```bash时h
./scripts/test_api.sh
```

## 配置说明

### 服务器配置
```yaml
server:
  port: 8199          # 服务端口
  mode: debug         # 运行模式: debug, release
  domain: http://localhost:8199  # 服务域名
```

### 日志配置
```yaml
logger:
  level: info         # 日志级别: debug, info, warn, error
  output: console     # 输出类型: console, file, both
  file:
    path: ./logs/transfer.log
    max_size: 100     # 单个文件最大尺寸(MB)
    max_age: 7        # 文件保留天数
    max_backups: 10   # 最大保留文件数
    compress: true    # 是否压缩
```

### 媒体转存配置
```yaml
media_transfer:
  download_timeout: 30      # 下载超时时间(秒)
  max_file_size: 104857600  # 最大文件大小(字节) 100MB
  retry_count: 3            # 重试次数
  allowed_domains: []       # 允许的域名列表，空表示允许所有
```

## 容错机制

1. **重试机制**: 下载失败时自动重试（默认3次）
2. **超时控制**: 下载超时保护（默认30秒）
3. **文件大小限制**: 防止下载过大文件（默认100MB）
4. **媒体类型验证**: 只允许上传支持的媒体格式
5. **URL格式验证**: 确保输入URL格式正确

## 开发命令

```bash
# 构建应用
make build

# 运行应用
make run

# 开发模式（生成文档并运行）
make dev

# 生成API文档
make docs

# 安装依赖
make deps

# 格式化代码
make fmt

# 运行测试
make test

# 清理构建文件
make clean
```

## 部署建议

### Docker部署
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o bin/transfer cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bin/transfer .
COPY --from=builder /app/config ./config
CMD ["./transfer"]
```

### 环境变量
生产环境建议通过环境变量 `TRANSFER_CONFIG` 指定配置文件：
```bash
export TRANSFER_CONFIG=production.yaml
```

## 注意事项

1. 请确保阿里云OSS配置信息正确
2. 建议在生产环境中使用HTTPS
3. 根据实际需求调整文件大小限制和超时时间
4. 定期清理OSS中的临时文件
5. 监控API调用频率，避免被目标网站限制
6. 生产环境建议设置日志输出为文件模式