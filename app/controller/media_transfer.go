package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"transfer/internal/dto"
	"transfer/internal/service"
	"transfer/pkg/config"
	"transfer/pkg/logger"
	"transfer/pkg/resp"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MediaTransferController struct {
	mediaTransferService *service.MediaTransferService
}

func NewMediaTransferController(cfg *config.Config) (*MediaTransferController, error) {
	mediaTransferService, err := service.NewMediaTransferService(cfg)
	if err != nil {
		return nil, err
	}

	return &MediaTransferController{
		mediaTransferService: mediaTransferService,
	}, nil
}

func (c *MediaTransferController) TransferMedia(ctx *gin.Context) {
	var req dto.MediaTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request parameters", zap.String("error", err.Error()))
		resp.Fail(ctx, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 清理URL中的空格
	req.URL = strings.TrimSpace(req.URL)

	// 基本验证
	if req.URL == "" {
		logger.Warn("URL cannot be empty")
		resp.Fail(ctx, http.StatusBadRequest, "URL cannot be empty", "")
		return
	}
	if req.PredictionUUID == "" {
		logger.Warn("PredictionUUID cannot be empty")
		resp.Fail(ctx, http.StatusBadRequest, "PredictionUUID cannot be empty", "")
		return
	}

	logger.Info("开始处理媒体转存请求",
		zap.String("url", req.URL),
		zap.String("ext", req.Ext),
		zap.String("prediction_uuid", req.PredictionUUID))

	// 验证URL格式和协议
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		resp.Fail(ctx, http.StatusBadRequest, "Invalid URL format", err.Error())
		return
	}

	// 验证协议必须是HTTP或HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		resp.Fail(ctx, http.StatusBadRequest, "Only HTTP and HTTPS protocols are supported", fmt.Sprintf("Unsupported protocol: %s", parsedURL.Scheme))
		return
	}

	// 执行转存
	ossURL, fileSize, contentType, err := c.mediaTransferService.TransferMedia(req.URL, req.Ext, req.PredictionUUID)
	if err != nil {
		logger.Error("媒体转存失败",
			zap.String("url", req.URL),
			zap.String("prediction_uuid", req.PredictionUUID),
			zap.String("error", err.Error()))
		resp.Fail(ctx, http.StatusInternalServerError, "Media transfer failed", err.Error())
		return
	}

	// 记录成功日志
	logger.Info("媒体转存成功",
		zap.String("url", req.URL),
		zap.String("prediction_uuid", req.PredictionUUID),
		zap.String("oss_url", ossURL),
		zap.Int64("file_size", fileSize),
		zap.String("content_type", contentType))

	// 返回成功响应
	response := dto.MediaTransferResponse{
		Success:     true,
		Message:     "Successfully uploaded file to OSS",
		OSSUrl:      ossURL,
		OriginalURL: req.URL,
		FileSize:    fileSize,
		ContentType: contentType,
	}

	resp.Success(ctx, response)
}

// Health 健康检查接口
// @Summary 健康检查
// @Description 文件转存服务健康检查
// @Tags 文件转存
// @Produce json
// @Success 200 {object} map[string]interface{} "服务正常"
// @Router /api/v1/media/health [get]
func (c *MediaTransferController) Health(ctx *gin.Context) {
	resp.Success(ctx, gin.H{
		"status":  "ok",
		"service": "media-transfer",
		"version": "1.0.0",
	})
}
