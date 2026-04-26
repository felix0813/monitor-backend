package handlers

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	authHandler := NewAuthHandler()
	commandHandler := NewCommandHandler()

	r.POST("/login", authHandler.Login)
	r.POST("/command", authHandler.AuthMiddleware(), commandHandler.ExecuteCommand)

	api := r.Group("/api")
	api.Use(authHandler.AuthMiddleware())
	{
		serviceHandler := NewServiceHandler()
		endpointHandler := NewEndpointHandler()
		navigationSiteHandler := NewNavigationSiteHandler()
		todoHandler := NewTodoHandler()
		commandTemplateHandler := NewCommandTemplateHandler()

		api.POST("/services", serviceHandler.CreateService)
		api.GET("/services", serviceHandler.ListServices)
		api.GET("/services/:id", serviceHandler.GetService)
		api.PUT("/services/:id", serviceHandler.UpdateService)
		api.DELETE("/services/:id", serviceHandler.DeleteService)

		api.POST("/services/:id/endpoints", endpointHandler.CreateEndpoint)
		api.GET("/services/:id/endpoints", endpointHandler.ListEndpoints)

		api.GET("/endpoints/:id", endpointHandler.GetEndpoint)
		api.PUT("/endpoints/:id", endpointHandler.UpdateEndpoint)
		api.DELETE("/endpoints/:id", endpointHandler.DeleteEndpoint)
		api.POST("/endpoints/:id/check", endpointHandler.CheckEndpointNow)

		api.GET("/navigation-sites", navigationSiteHandler.ListNavigationSites)
		api.POST("/navigation-sites", navigationSiteHandler.CreateNavigationSite)
		api.PUT("/navigation-sites/:id", navigationSiteHandler.UpdateNavigationSite)
		api.DELETE("/navigation-sites/:id", navigationSiteHandler.DeleteNavigationSite)
		api.PUT("/navigation-sites/order", navigationSiteHandler.ReorderNavigationSites)

		api.GET("/todo-projects", todoHandler.ListTodoProjects)
		api.POST("/todo-projects", todoHandler.CreateTodoProject)
		api.PUT("/todo-projects/:project_id", todoHandler.UpdateTodoProject)
		api.DELETE("/todo-projects/:project_id", todoHandler.DeleteTodoProject)
		api.GET("/todo-projects/:project_id/items", todoHandler.ListTodoItems)
		api.POST("/todo-projects/:project_id/items", todoHandler.CreateTodoItem)
		api.PUT("/todo-projects/:project_id/items/:item_id", todoHandler.UpdateTodoItem)
		api.DELETE("/todo-projects/:project_id/items/:item_id", todoHandler.DeleteTodoItem)

		api.POST("/command-templates", commandTemplateHandler.CreateCommandTemplate)
		api.GET("/command-templates", commandTemplateHandler.ListCommandTemplates)
		api.GET("/command-templates/:id", commandTemplateHandler.GetCommandTemplate)
		api.PUT("/command-templates/:id", commandTemplateHandler.UpdateCommandTemplate)
		api.DELETE("/command-templates/:id", commandTemplateHandler.DeleteCommandTemplate)
		api.POST("/command", commandHandler.ExecuteCommand)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
