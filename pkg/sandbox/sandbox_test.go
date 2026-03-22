package sandbox

import (
	"path/filepath"
	"testing"
)

func TestPathTranslator_ToPhysical(t *testing.T) {
	threadData := &ThreadData{
		WorkspacePath: "/real/path/workspace",
		UploadsPath:   "/real/path/uploads",
		OutputsPath:   "/real/path/outputs",
	}

	translator := NewPathTranslator(threadData)

	tests := []struct {
		name        string
		virtualPath string
		expected    string
		expectErr   bool
	}{
		{
			name:        "workspace path",
			virtualPath: "/mnt/user-data/workspace/file.txt",
			expected:    "/real/path/workspace/file.txt",
			expectErr:   false,
		},
		{
			name:        "uploads path",
			virtualPath: "/mnt/user-data/uploads/image.png",
			expected:    "/real/path/uploads/image.png",
			expectErr:   false,
		},
		{
			name:        "outputs path",
			virtualPath: "/mnt/user-data/outputs/report.pdf",
			expected:    "/real/path/outputs/report.pdf",
			expectErr:   false,
		},
		{
			name:        "path traversal",
			virtualPath: "/mnt/user-data/workspace/../secret.txt",
			expected:    "",
			expectErr:   true,
		},
		{
			name:        "outside sandbox",
			virtualPath: "/etc/passwd",
			expected:    "",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := translator.ToPhysical(tt.virtualPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("ToPhysical() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr {
				if filepath.Clean(result) != filepath.Clean(tt.expected) {
					t.Errorf("ToPhysical() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestPathTranslator_ToVirtual(t *testing.T) {
	threadData := &ThreadData{
		WorkspacePath: "/real/path/workspace",
		UploadsPath:   "/real/path/uploads",
		OutputsPath:   "/real/path/outputs",
	}

	translator := NewPathTranslator(threadData)

	tests := []struct {
		name         string
		physicalPath string
		expected     string
	}{
		{
			name:         "workspace path",
			physicalPath: "/real/path/workspace/file.txt",
			expected:     "/mnt/user-data/workspace/file.txt",
		},
		{
			name:         "uploads path",
			physicalPath: "/real/path/uploads/image.png",
			expected:     "/mnt/user-data/uploads/image.png",
		},
		{
			name:         "outputs path",
			physicalPath: "/real/path/outputs/report.pdf",
			expected:     "/mnt/user-data/outputs/report.pdf",
		},
		{
			name:         "non-sandbox path",
			physicalPath: "/other/path/file.txt",
			expected:     "/other/path/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.ToVirtual(tt.physicalPath)
			if result != tt.expected {
				t.Errorf("ToVirtual() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPathTranslator_ValidatePath(t *testing.T) {
	threadData := &ThreadData{
		WorkspacePath: "/real/path/workspace",
	}

	translator := NewPathTranslator(threadData)

	tests := []struct {
		name          string
		virtualPath   string
		allowOutside  bool
		expectErr     bool
	}{
		{
			name:         "valid workspace path",
			virtualPath:  "/mnt/user-data/workspace/file.txt",
			allowOutside: false,
			expectErr:    false,
		},
		{
			name:         "valid uploads path",
			virtualPath:  "/mnt/user-data/uploads/file.txt",
			allowOutside: false,
			expectErr:    false,
		},
		{
			name:         "valid outputs path",
			virtualPath:  "/mnt/user-data/outputs/file.txt",
			allowOutside: false,
			expectErr:    false,
		},
		{
			name:         "path traversal",
			virtualPath:  "/mnt/user-data/workspace/../secret.txt",
			allowOutside: false,
			expectErr:    true,
		},
		{
			name:         "outside sandbox - not allowed",
			virtualPath:  "/etc/passwd",
			allowOutside: false,
			expectErr:    true,
		},
		{
			name:         "outside sandbox - allowed",
			virtualPath:  "/etc/passwd",
			allowOutside: true,
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := translator.ValidatePath(tt.virtualPath, tt.allowOutside)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidatePath() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestIsVirtualPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "workspace path",
			path:     "/mnt/user-data/workspace/file.txt",
			expected: true,
		},
		{
			name:     "uploads path",
			path:     "/mnt/user-data/uploads/file.txt",
			expected: true,
		},
		{
			name:     "outputs path",
			path:     "/mnt/user-data/outputs/file.txt",
			expected: true,
		},
		{
			name:     "skills path",
			path:     "/mnt/skills/skill.md",
			expected: true,
		},
		{
			name:     "non-virtual path",
			path:     "/real/path/file.txt",
			expected: false,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVirtualPath(tt.path)
			if result != tt.expected {
				t.Errorf("IsVirtualPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTranslatePathsInCommand(t *testing.T) {
	threadData := &ThreadData{
		WorkspacePath: "/real/path/workspace",
		UploadsPath:   "/real/path/uploads",
		OutputsPath:   "/real/path/outputs",
	}

	tests := []struct {
		name        string
		command     string
		expected    string
	}{
		{
			name:        "workspace path in command",
			command:     "ls /mnt/user-data/workspace",
			expected:    "ls /real/path/workspace",
		},
		{
			name:        "multiple paths",
			command:     "cp /mnt/user-data/uploads/file.txt /mnt/user-data/outputs/",
			expected:    "cp /real/path/uploads/file.txt /real/path/outputs/",
		},
		{
			name:        "no virtual paths",
			command:     "echo hello",
			expected:    "echo hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslatePathsInCommand(tt.command, threadData, "")
			if result != tt.expected {
				t.Errorf("TranslatePathsInCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}
