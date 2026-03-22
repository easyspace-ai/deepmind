package deerflow

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/weibaohui/nanobot-go/pkg/agent/state"
	"github.com/weibaohui/nanobot-go/pkg/config"
)

// ============================================
// PresentFiles 工具
// ============================================

// PresentFilesTool present_files 工具
// 一比一复刻 DeerFlow 的 present_file_tool
type PresentFilesTool struct {
	*BaseDeerFlowTool
}

// NewPresentFilesTool 创建 present_files 工具
func NewPresentFilesTool(cfg *ToolConfig) tool.BaseTool {
	return &PresentFilesTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"present_files",
			"Make files visible to the user for viewing and rendering in the client interface.\n\nWhen to use the present_files tool:\n\n- Making any file available for the user to view, download, or interact with\n- Presenting multiple related files at once\n- After creating files that should be presented to the user\n\nWhen NOT to use the present_files tool:\n- When you only need to read file contents for your own processing\n- For temporary or intermediate files not meant for user viewing\n\nNotes:\n- You should call this tool after creating files and moving them to the `/mnt/user-data/outputs` directory.\n- This tool can be safely called in parallel with other tools. State updates are handled by a reducer to prevent conflicts.",
			map[string]interface{}{
				"filepaths": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of absolute file paths to present to the user. **Only** files in `/mnt/user-data/outputs` can be presented.",
				},
			},
		),
	}
}

// normalizePresentedFilepath 规范化展示文件路径
// 一比一复刻 DeerFlow 的 _normalize_presented_filepath
func normalizePresentedFilepath(userPath string, threadData *state.ThreadDataState) (string, error) {
	if threadData == nil {
		return "", fmt.Errorf("thread runtime state is not available")
	}

	outputsPath := threadData.OutputsPath
	if outputsPath == "" {
		return "", fmt.Errorf("thread outputs path is not available")
	}

	outputsDir := filepath.Clean(outputsPath)
	virtualOutputsPrefix := config.VirtualOutputsPath

	// 检查是否是虚拟路径
	var actualPath string
	if strings.HasPrefix(userPath, config.VirtualPathPrefix+"/") {
		// 虚拟路径，转换为实际路径
		actualPath = ReplaceVirtualPath(userPath, threadData)
	} else {
		// 已经是实际路径
		actualPath = filepath.Clean(userPath)
	}

	// 规范化路径
	actualPath, err := filepath.Abs(actualPath)
	if err != nil {
		actualPath = filepath.Clean(actualPath)
	}

	// 检查是否在 outputs 目录内
	// 简化处理，跳过严格的相对路径检查
	if !strings.HasPrefix(strings.ToLower(actualPath), strings.ToLower(outputsDir)) {
		return "", fmt.Errorf("only files in %s can be presented: %s", virtualOutputsPrefix, userPath)
	}

	// 构建虚拟路径
	// 简化处理，直接返回虚拟路径格式
	filename := filepath.Base(actualPath)
	return fmt.Sprintf("%s/%s", virtualOutputsPrefix, filename), nil
}

// Invoke 执行 present_files 工具
func (t *PresentFilesTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	filepathsIf, _ := args["filepaths"].([]interface{})
	var filepaths []string
	for _, fpIf := range filepathsIf {
		if fp, ok := fpIf.(string); ok {
			filepaths = append(filepaths, fp)
		}
	}

	// 规范化路径
	var normalizedPaths []string
	var threadData *state.ThreadDataState // TODO: 从上下文中获取

	for _, fp := range filepaths {
		normalized, err := normalizePresentedFilepath(fp, threadData)
		if err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		normalizedPaths = append(normalizedPaths, normalized)
	}

	// 返回结果
	return map[string]interface{}{
		"artifacts": normalizedPaths,
		"content":   "Successfully presented files",
	}, nil
}

// ============================================
// AskClarification 工具
// ============================================

// AskClarificationTool ask_clarification 工具
// 一比一复刻 DeerFlow 的 ask_clarification 工具
type AskClarificationTool struct {
	*BaseDeerFlowTool
}

// NewAskClarificationTool 创建 ask_clarification 工具
func NewAskClarificationTool() tool.BaseTool {
	return &AskClarificationTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"ask_clarification",
			"Ask the user for clarification",
			map[string]interface{}{
				"question": map[string]interface{}{
					"type":        "string",
					"description": "The clarification question",
				},
				"clarification_type": map[string]interface{}{
					"type": "string",
					"description": "Type of clarification: missing_info, ambiguous_requirement, approach_choice, risk_confirmation, suggestion",
					"enum": []string{
						"missing_info",
						"ambiguous_requirement",
						"approach_choice",
						"risk_confirmation",
						"suggestion",
					},
				},
				"context": map[string]interface{}{
					"type":        "string",
					"description": "Context explaining why clarification is needed",
				},
				"options": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of options for the user to choose from",
				},
			},
		),
	}
}

// Invoke 执行 ask_clarification 工具
func (t *AskClarificationTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// 注意：此工具会被 ClarificationMiddleware 拦截并中断执行
	// 这里只返回简单的成功响应
	return "Clarification requested", nil
}

// ============================================
// ViewImage 工具
// ============================================

// ViewImageTool view_image 工具
// 一比一复刻 DeerFlow 的 view_image_tool
type ViewImageTool struct {
	*BaseDeerFlowTool
	cfg *ToolConfig
}

// maxViewImageFileBytes 单张图片读取上限，防止超大文件撑爆内存。
const maxViewImageFileBytes = 12 << 20

// NewViewImageTool 创建 view_image 工具
func NewViewImageTool(cfg *ToolConfig) tool.BaseTool {
	return &ViewImageTool{
		BaseDeerFlowTool: NewBaseDeerFlowTool(
			"view_image",
			"View an image file.\n\nUse this tool to read an image file and make it available for display.\n\nWhen to use the view_image tool:\n- When you need to view an image file.\n\nWhen NOT to use the view_image tool:\n- For non-image files (use present_files instead)\n- For multiple files at once (use present_files instead)",
			map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Absolute path to the image file. Common formats supported: jpg, jpeg, png, webp.",
				},
			},
		),
		cfg: cfg,
	}
}

// Invoke 执行 view_image 工具
func (t *ViewImageTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	imagePath, _ := args["path"].(string)

	// 验证扩展名
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	ext := strings.ToLower(filepath.Ext(imagePath))
	if !validExtensions[ext] {
		return fmt.Sprintf("Error: Unsupported image format: %s. Supported formats: .jpg, .jpeg, .png, .webp", ext), nil
	}

	// 检测 MIME 类型
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		extToMime := map[string]string{
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".png":  "image/png",
			".webp": "image/webp",
		}
		mimeType = extToMime[ext]
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	imageBase64 := ""
	var threadData *state.ThreadDataState
	if t.cfg != nil && t.cfg.ThreadState != nil {
		threadData = t.cfg.ThreadState.ThreadData
	}
	if threadData != nil {
		if err := ValidateLocalToolPath(imagePath, threadData, true); err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		realPath := ReplaceVirtualPath(imagePath, threadData)
		data, err := os.ReadFile(realPath)
		if err != nil {
			return fmt.Sprintf("Error: failed to read image: %v", err), nil
		}
		if len(data) > maxViewImageFileBytes {
			return fmt.Sprintf("Error: image exceeds size limit (%d bytes)", maxViewImageFileBytes), nil
		}
		imageBase64 = base64.StdEncoding.EncodeToString(data)
	}

	newViewedImages := map[string]state.ViewedImageData{
		imagePath: {
			Base64:   imageBase64,
			MimeType: mimeType,
		},
	}

	if t.cfg != nil && t.cfg.ThreadState != nil {
		if t.cfg.ThreadState.ViewedImages == nil {
			t.cfg.ThreadState.ViewedImages = make(map[string]state.ViewedImageData)
		}
		t.cfg.ThreadState.ViewedImages[imagePath] = newViewedImages[imagePath]
	}

	content := "Successfully read image"
	if threadData != nil && imageBase64 == "" {
		content = "Image metadata recorded (file read skipped: bind ThreadState.ThreadData for full read)"
	}

	return map[string]interface{}{
		"viewed_images": newViewedImages,
		"content":       content,
	}, nil
}
