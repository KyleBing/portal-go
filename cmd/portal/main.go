package main

import (
	"fmt"
	"log"
	"os"

	"github.com/KyleBing/portal-go/internal/config"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/server"
	"github.com/KyleBing/portal-go/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate()
			return
		case "serve":
			runServe()
			return
		case "version":
			fmt.Println(version.Version)
			return
		case "help", "-h", "--help":
			usage()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			usage()
			os.Exit(1)
		}
	}
	runServe()
}

func usage() {
	fmt.Fprintf(os.Stderr, `portal — Portal Go API

Usage:
  portal              start HTTP server (default)
  portal serve        start HTTP server
  portal migrate      apply embedded diary schema migrations
  portal version      print version
`)
}

func runMigrate() {
	if _, err := config.Load(); err != nil {
		log.Fatalf("加载数据库配置失败: %v", err)
	}
	defer db.CloseAll()

	diary, err := db.Open(db.Diary)
	if err != nil {
		log.Fatalf("连接 diary 失败: %v", err)
	}
	if err := db.Migrate(diary); err != nil {
		log.Fatalf("migrate 失败: %v", err)
	}
	log.Println("migrate 完成")
}

func runServe() {
	// 加载数据库配置（未初始化时允许失败，交由安装引导流程处理）
	if _, err := config.Load(); err != nil {
		log.Printf("加载数据库配置失败（可能尚未完成安装引导）: %v", err)
	}
	defer db.CloseAll()

	r := server.New()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Portal 服务启动 v%s，监听端口 :%s", version.Version, port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
