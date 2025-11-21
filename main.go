package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monitor/db"
	"monitor/handlers"
	"monitor/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化 MongoDB
	err := db.InitMongo()
	if err != nil {
		log.Fatal("Failed to connect MongoDB:", err)
	}
	log.Println("MongoDB connected")

	// 2. 启动定时检查器
	scheduler.StartChecker()
	log.Println("Checker started")

	// 3. 初始化 Gin
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	// 4. 注册路由（示例）
	handlers.RegisterRoutes(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// 5. 启动服务器（支持优雅退出）
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Println("Server running ")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %s\n", err)
		}
	}()

	// 6. 优雅退出处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
