package handlers

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	serviceHandler := NewServiceHandler()
	endpointHandler := NewEndpointHandler()

	api := r.Group("/api")

	// 服务管理
	api.POST("/services", serviceHandler.CreateService)
	api.GET("/services", serviceHandler.ListServices)
	api.GET("/services/:id", serviceHandler.GetService)
	api.PUT("/services/:id", serviceHandler.UpdateService)
	api.DELETE("/services/:id", serviceHandler.DeleteService)

	// 端点管理
	api.POST("/services/:id/endpoints", endpointHandler.CreateEndpoint)
	api.GET("/services/:id/endpoints", endpointHandler.ListEndpoints)

	api.GET("/endpoints/:id", endpointHandler.GetEndpoint)
	api.PUT("/endpoints/:id", endpointHandler.UpdateEndpoint)
	api.DELETE("/endpoints/:id", endpointHandler.DeleteEndpoint)

	api.POST("/endpoints/:id/check", endpointHandler.CheckEndpointNow)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
