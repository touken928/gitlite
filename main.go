package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"gitlite/internal/server"
)

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func main() {
	// 从环境变量读取配置，支持命令行参数（可选）
	port := getEnv("GITLITE_PORT", "2222")
	dataPath := getEnv("GITLITE_DATA", "data")

	srv, err := server.New(port, dataPath)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	log.Printf("Git 服务器已启动，监听端口 %s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
	srv.Stop()
}
