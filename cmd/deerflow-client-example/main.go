package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/weibaohui/nanobot-go/pkg/deerflow"
)

func main() {
	fmt.Println("🦌 DeerFlow Go Client 示例")
	fmt.Println("==============================")

	// 创建客户端
	client := deerflow.NewClient(
		deerflow.WithBaseURL("http://localhost:8001"),
	)

	// 设置上下文，处理信号
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n收到终止信号，正在退出...")
		cancel()
	}()

	// 示例 1: 列出可用模型
	fmt.Println("\n1. 列出可用模型")
	fmt.Println("------------------")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Printf("⚠️  获取模型失败 (后端未启动?): %v", err)
	} else {
		for _, model := range models {
			thinking := ""
			if model.SupportsThinking {
				thinking = " (支持思维链)"
			}
			fmt.Printf("  • %s%s\n", model.DisplayName, thinking)
		}
	}

	// 示例 2: 列出可用技能
	fmt.Println("\n2. 列出可用技能")
	fmt.Println("------------------")
	skills, err := client.ListSkills(ctx)
	if err != nil {
		log.Printf("⚠️  获取技能失败: %v", err)
	} else {
		for _, skill := range skills {
			status := "✓"
			if !skill.Enabled {
				status = "✗"
			}
			fmt.Printf("  %s %s\n", status, skill.Name)
		}
	}

	// 示例 3: 简单对话 (非流式)
	fmt.Println("\n3. 简单对话 (非流式)")
	fmt.Println("---------------------")
	chatCtx, chatCancel := context.WithTimeout(ctx, 60*time.Second)
	defer chatCancel()

	response, err := client.Chat(chatCtx, "你好！请用一句话介绍一下自己。")
	if err != nil {
		log.Printf("⚠️  对话失败: %v", err)
		log.Println("提示: 请确保后端服务正在运行: go run ./cmd/nanobot gateway")
	} else {
		fmt.Printf("AI: %s\n", response.Content)
		fmt.Printf("(Thread ID: %s)\n", response.ThreadID)
	}

	// 示例 4: 流式对话
	fmt.Println("\n4. 流式对话")
	fmt.Println("-------------")
	streamCtx, streamCancel := context.WithTimeout(ctx, 60*time.Second)
	defer streamCancel()

	fmt.Print("AI: ")
	stream, err := client.Stream(streamCtx, "给我讲一个关于程序员的短故事")
	if err != nil {
		log.Printf("⚠️  流式对话失败: %v", err)
	} else {
		defer stream.Close()

		for event := range stream.Events() {
			switch e := event.(type) {
			case *deerflow.MessageEvent:
				fmt.Print(e.Content)
			case *deerflow.MetadataEvent:
				if e.RunID != "" {
					fmt.Printf("\n[Run ID: %s]\n", e.RunID)
					fmt.Print("AI: ")
				}
			case *deerflow.FinishEvent:
				fmt.Printf("\n[完成: %s]\n", e.Status)
			}
		}

		if err := stream.Err(); err != nil {
			log.Printf("\n⚠️  流错误: %v", err)
		}
	}

	// 示例 5: 列出历史线程
	fmt.Println("\n5. 历史线程")
	fmt.Println("-------------")
	threads, err := client.ListThreads(ctx, &deerflow.ListThreadsRequest{
		Limit:     5,
		SortBy:    "updated_at",
		SortOrder: "desc",
	})
	if err != nil {
		log.Printf("⚠️  获取线程失败: %v", err)
	} else if len(threads) == 0 {
		fmt.Println("  (暂无历史线程)")
	} else {
		for _, thread := range threads {
			fmt.Printf("  • %s (创建于: %s)\n", thread.ThreadID, thread.CreatedAt)
		}
	}

	fmt.Println("\n==============================")
	fmt.Println("示例完成！")
	fmt.Println("\n更多用法请参考:")
	fmt.Println("  - pkg/deerflow/example_test.go")
	fmt.Println("  - pkg/deerflow/client.go")
}
