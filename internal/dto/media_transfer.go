package dto

// MediaTransferRequest 文件转存请求
type MediaTransferRequest struct {
	URL            string `json:"url" binding:"required"`
	Ext            string `json:"ext" binding:"required"`
	PredictionUUID string `json:"prediction_uuid" binding:"required"`
}

// MediaTransferResponse 文件转存响应
type MediaTransferResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	OSSUrl      string `json:"oss_url"`
	OriginalURL string `json:"original_url"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
}
