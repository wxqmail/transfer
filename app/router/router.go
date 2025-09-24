package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"transfer/app/controller"
	"transfer/app/middleware"
	"transfer/pkg/config"
	"transfer/pkg/resp"

	_ "transfer/docs" // 导入生成的swagger文档
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config, mediaTransferController *controller.MediaTransferController) (*gin.Engine, error) {
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间件
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// 根路径
	r.GET("/", func(c *gin.Context) {
		resp.Success(c, gin.H{
			"service":     "File Transfer API",
			"version":     "1.0.0",
			"description": "支持所有文件类型的转存服务，无文件大小限制",
			"endpoints": gin.H{
				"health":   "GET /api/v1/media/health",
				"transfer": "POST /api/v1/media/transfer",
				"docs":     "GET /swagger/index.html",
			},
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		media := api.Group("/media")
		{
			media.GET("/health", mediaTransferController.Health)
			media.POST("/transfer", mediaTransferController.TransferMedia)
		}
	}

	// Swagger文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 404处理器
	r.NoRoute(func(c *gin.Context) {
		resp.Fail(c, http.StatusNotFound, "Route not found", "The requested endpoint does not exist")
	})

	return r, nil
}
