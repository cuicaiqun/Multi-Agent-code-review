package main

import (
	"log"
	"os"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/handler"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8082"
	}

	router := handler.SetupRouter()
	log.Printf("Go版合同审查系统启动在 :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
