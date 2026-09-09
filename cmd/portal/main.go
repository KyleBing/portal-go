package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if _, err := config.Load(); err != nil {
		log.Printf("加载数据库配置失败（可能尚未完成安装引导）: %v", err)
	}
	defer db.CloseAll()

	r := server.New()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Portal 服务启动 v%s，监听端口 :%s", version.Version, port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("正在优雅关闭 Portal …")
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("关闭异常: %v", err)
	}
}
