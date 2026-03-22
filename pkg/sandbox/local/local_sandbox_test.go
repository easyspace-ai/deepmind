package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/weibaohui/nanobot-go/pkg/sandbox"
)

func TestLocalSandbox(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "sandbox-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	threadData := &sandbox.ThreadData{
		WorkspacePath: filepath.Join(tmpDir, "workspace"),
		UploadsPath:   filepath.Join(tmpDir, "uploads"),
		OutputsPath:   filepath.Join(tmpDir, "outputs"),
	}

	// 创建目录
	for _, dir := range []string{threadData.WorkspacePath, threadData.UploadsPath, threadData.OutputsPath} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	sb := NewLocalSandbox("local", threadData)

	// 测试 WriteFile
	testFile := "/mnt/user-data/workspace/test.txt"
	err = sb.WriteFile(testFile, "hello world", false)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// 测试 ReadFile
	content, err := sb.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if content != "hello world" {
		t.Errorf("ReadFile() = %v, want 'hello world'", content)
	}

	// 测试 Append
	err = sb.WriteFile(testFile, " appended", true)
	if err != nil {
		t.Fatalf("WriteFile(append) error = %v", err)
	}
	content, err = sb.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if content != "hello world appended" {
		t.Errorf("ReadFile() after append = %v, want 'hello world appended'", content)
	}

	// 测试 ExecuteCommand
	output, err := sb.ExecuteCommand("echo 'test command'")
	if err != nil {
		// 某些环境可能没有 bash，不视为失败
		t.Logf("ExecuteCommand() error = %v (may not have bash)", err)
	} else {
		if output == "" {
			t.Log("ExecuteCommand() returned empty output")
		}
	}

	// 测试 ListDir
	_, err = sb.ListDir("/mnt/user-data/workspace", 2)
	if err != nil {
		t.Fatalf("ListDir() error = %v", err)
	}
}

func TestLocalSandboxProvider(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "provider-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	provider := GetLocalSandboxProvider(tmpDir)
	defer provider.Reset()

	// 测试 Acquire
	threadID := "test-thread-123"
	sb, err := provider.Acquire(threadID)
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if sb == nil {
		t.Fatal("Acquire() returned nil")
	}
	if sb.ID() != "local" {
		t.Errorf("Sandbox ID = %v, want 'local'", sb.ID())
	}

	// 测试 Get
	sb2, exists := provider.Get(threadID)
	if !exists {
		t.Fatal("Get() should return true")
	}
	if sb2 == nil {
		t.Fatal("Get() returned nil")
	}

	// 测试再次 Acquire 应该返回同一个实例
	sb3, err := provider.Acquire(threadID)
	if err != nil {
		t.Fatalf("Acquire() again error = %v", err)
	}
	if sb3 != sb {
		t.Error("Acquire() should return the same instance")
	}

	// 测试 Release
	err = provider.Release(threadID)
	if err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	// Release 后应该不存在
	_, exists = provider.Get(threadID)
	if exists {
		t.Error("Get() after Release should return false")
	}
}
