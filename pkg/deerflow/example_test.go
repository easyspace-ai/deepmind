package deerflow_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/weibaohui/nanobot-go/pkg/deerflow"
)

func ExampleClient_Chat() {
	// Create a client
	client := deerflow.NewClient(
		deerflow.WithBaseURL("http://localhost:8001"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simple chat
	response, err := client.Chat(ctx, "Hello, what can you do?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", response.Content)
	fmt.Printf("Thread ID: %s\n", response.ThreadID)
}

func ExampleClient_Chat_withThread() {
	client := deerflow.NewClient(
		deerflow.WithBaseURL("http://localhost:8001"),
	)

	ctx := context.Background()

	// Create a thread first
	thread, err := client.CreateThread(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Chat with the thread
	response, err := client.Chat(ctx, "Tell me a joke",
		deerflow.WithThreadID(thread.ThreadID))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Joke: %s\n", response.Content)
}

func ExampleClient_Stream() {
	client := deerflow.NewClient(
		deerflow.WithBaseURL("http://localhost:8001"),
	)

	ctx := context.Background()

	// Start streaming
	stream, err := client.Stream(ctx, "Tell me a story about a robot")
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	// Read events
	for event := range stream.Events() {
		switch e := event.(type) {
		case *deerflow.MessageEvent:
			fmt.Print(e.Content)
		case *deerflow.ToolCallEvent:
			fmt.Printf("\n[Tool call: %s]\n", e.Name)
		case *deerflow.ToolResultEvent:
			fmt.Printf("\n[Tool result: %s]\n", e.Name)
		case *deerflow.MetadataEvent:
			fmt.Printf("\n[Metadata: run_id=%s, thread_id=%s]\n", e.RunID, e.ThreadID)
		case *deerflow.FinishEvent:
			fmt.Printf("\n[Finished: %s]\n", e.Status)
		}
	}

	// Check for errors
	if err := stream.Err(); err != nil {
		log.Printf("Stream error: %v", err)
	}
}

func ExampleClient_ListModels() {
	client := deerflow.NewClient(
		deerflow.WithBaseURL("http://localhost:8001"),
	)

	ctx := context.Background()

	// List available models
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Available models:")
	for _, model := range models {
		fmt.Printf("  - %s (%s)\n", model.DisplayName, model.ID)
	}
}

func ExampleClient_ListThreads() {
	client := deerflow.NewClient(
		deerflow.WithBaseURL("http://localhost:8001"),
	)

	ctx := context.Background()

	// List recent threads
	threads, err := client.ListThreads(ctx, &deerflow.ListThreadsRequest{
		Limit:     10,
		SortBy:    "updated_at",
		SortOrder: "desc",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Recent threads:")
	for _, thread := range threads {
		fmt.Printf("  - %s (updated: %s)\n", thread.ThreadID, thread.UpdatedAt)
	}
}
