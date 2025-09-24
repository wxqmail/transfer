package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"transfer/pkg/config"
	"transfer/pkg/logger"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"go.uber.org/zap"
)

type MediaTransferService struct {
	ossClient *oss.Client
	bucket    *oss.Bucket
	config    *config.Config
}

type DownloadResult struct {
	Reader      io.ReadCloser
	ContentType string
	Size        int64
}

func NewMediaTransferService(cfg *config.Config) (*MediaTransferService, error) {
	// 创建OSS客户端
	endpoint := fmt.Sprintf("https://%s", cfg.AliyunOSS.Endpoint)

	client, err := oss.New(
		endpoint,
		cfg.AliyunOSS.AccessKeyID,
		cfg.AliyunOSS.AccessKeySecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取bucket
	bucket, err := client.Bucket(cfg.AliyunOSS.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	return &MediaTransferService{
		ossClient: client,
		bucket:    bucket,
		config:    cfg,
	}, nil
}

// TransferMedia 转存媒体文件
func (s *MediaTransferService) TransferMedia(mediaURL string, ext string, predictionUUID string) (string, int64, string, error) {
	// 清理URL中的空格
	mediaURL = strings.TrimSpace(mediaURL)

	// 下载文件
	result, err := s.download(mediaURL)
	if err != nil {
		logger.Error("文件下载失败", zap.String("url", mediaURL), zap.String("error", err.Error()))
		return "", 0, "", fmt.Errorf("download failed: %w", err)
	}
	defer result.Reader.Close()

	// 生成OSS对象键
	objectKey := s.generateObjectKey(mediaURL, result.ContentType, ext, predictionUUID)

	// 上传到OSS
	ossURL, err := s.uploadToOSS(result.Reader, objectKey, result.ContentType)
	if err != nil {
		logger.Error("OSS上传失败", zap.String("object_key", objectKey), zap.String("error", err.Error()))
		return "", 0, "", fmt.Errorf("upload to OSS failed: %w", err)
	}

	logger.Info("OSS上传成功", zap.String("oss_url", ossURL))
	return ossURL, result.Size, result.ContentType, nil
}

// download 下载媒体文件
func (s *MediaTransferService) download(mediaURL string) (*DownloadResult, error) {
	// 清理URL中的空格
	mediaURL = strings.TrimSpace(mediaURL)

	// 验证URL格式
	if _, err := url.Parse(mediaURL); err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	// 创建HTTP客户端
	client := &http.Client{}

	// 创建请求
	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 获取Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &DownloadResult{
		Reader:      resp.Body,
		ContentType: contentType,
		Size:        resp.ContentLength,
	}, nil
}

// uploadToOSS 上传文件到OSS
func (s *MediaTransferService) uploadToOSS(reader io.Reader, objectKey string, contentType string) (string, error) {
	// 设置上传选项，包括公共读取权限
	options := []oss.Option{
		oss.ContentType(contentType),
		oss.Meta("upload-time", time.Now().Format(time.RFC3339)),
		oss.ObjectACL(oss.ACLPublicRead), // 设置为公共读取权限，用户可以随意下载
	}

	// 上传文件
	err := s.bucket.PutObject(objectKey, reader, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// 生成公开访问URL
	url := fmt.Sprintf("https://%s.%s/%s", s.config.AliyunOSS.Bucket, s.config.AliyunOSS.Endpoint, objectKey)
	return url, nil
}

// generateObjectKey 生成对象键名
func (s *MediaTransferService) generateObjectKey(originalURL string, contentType string, ext string, predictionUUID string) string {
	var filename string

	// 解析URL以正确提取文件名
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		// 如果解析失败，使用简单的分割方法
		parts := strings.Split(originalURL, "/")
		filename = parts[len(parts)-1]
		// 移除查询参数
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
	} else {
		// 从路径中提取文件名，忽略查询参数
		pathParts := strings.Split(parsedURL.Path, "/")
		filename = pathParts[len(pathParts)-1]
	}

	// 如果文件名为空，生成一个默认名称
	if filename == "" {
		filename = "file"
	}

	// 优先使用传入的扩展名，否则根据content-type或文件名推断
	if ext != "" {
		// 确保扩展名以点开头
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		// 移除原有扩展名并添加新的
		if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 {
			filename = filename[:dotIndex]
		}
		filename += ext
	} else if !strings.Contains(filename, ".") {
		// 如果没有传入扩展名且文件名中没有扩展名，根据content-type添加
		autoExt := s.getExtensionFromContentType(contentType)
		if autoExt != "" {
			filename += autoExt
		}
	}

	// 使用 outputs/{PredictionUUID}/filename 格式
	objectKey := fmt.Sprintf("outputs/%s/%s", predictionUUID, filename)

	return objectKey
}

// getExtensionFromContentType 根据content-type获取文件扩展名
func (s *MediaTransferService) getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "image/jpeg"):
		return ".jpg"
	case strings.Contains(contentType, "image/png"):
		return ".png"
	case strings.Contains(contentType, "image/gif"):
		return ".gif"
	case strings.Contains(contentType, "image/webp"):
		return ".webp"
	case strings.Contains(contentType, "video/mp4"):
		return ".mp4"
	case strings.Contains(contentType, "video/avi"):
		return ".avi"
	case strings.Contains(contentType, "video/mov"):
		return ".mov"
	case strings.Contains(contentType, "audio/mp3"):
		return ".mp3"
	case strings.Contains(contentType, "audio/wav"):
		return ".wav"
	case strings.Contains(contentType, "audio/aac"):
		return ".aac"
	default:
		return ""
	}
}

// 移除媒体类型检查，支持所有文件类型
