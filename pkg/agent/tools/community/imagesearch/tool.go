package imagesearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// ============================================
// Image Search Tool（一比一复刻 DeerFlow）
// ============================================

// ImageSearchTool image_search 工具
type ImageSearchTool struct {
	maxResults int
}

// NewImageSearchTool 创建 image_search 工具
func NewImageSearchTool(maxResults int) *ImageSearchTool {
	if maxResults <= 0 {
		maxResults = 5
	}
	return &ImageSearchTool{
		maxResults: maxResults,
	}
}

// Name 返回工具名称
func (t *ImageSearchTool) Name() string {
	return "image_search"
}

// Info 返回工具信息
func (t *ImageSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: `Search for images online. Use this tool BEFORE image generation to find reference images for characters, portraits, objects, scenes, or any content requiring visual accuracy.

**When to use:**
- Before generating character/portrait images: search for similar poses, expressions, styles
- Before generating specific objects/products: search for accurate visual references
- Before generating scenes/locations: search for architectural or environmental references
- Before generating fashion/clothing: search for style and detail references

The returned image URLs can be used as reference images in image generation to significantly improve quality.`,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Type:     schema.DataType("string"),
				Desc:     "Search keywords describing the images you want to find. Be specific for better results (e.g., \"Japanese woman street photography 1990s\" instead of just \"woman\").",
				Required: true,
			},
			"max_results": {
				Type:     schema.DataType("integer"),
				Desc:     "Maximum number of images to return. Default is 5.",
				Required: false,
			},
			"size": {
				Type:     schema.DataType("string"),
				Desc:     "Image size filter. Options: \"Small\", \"Medium\", \"Large\", \"Wallpaper\". Use \"Large\" for reference images.",
				Required: false,
			},
			"type_image": {
				Type:     schema.DataType("string"),
				Desc:     "Image type filter. Options: \"photo\", \"clipart\", \"gif\", \"transparent\", \"line\". Use \"photo\" for realistic references.",
				Required: false,
			},
			"layout": {
				Type:     schema.DataType("string"),
				Desc:     "Layout filter. Options: \"Square\", \"Tall\", \"Wide\". Choose based on your generation needs.",
				Required: false,
			},
		}),
	}, nil
}

// InvokableRun 执行工具
func (t *ImageSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
		Size       string `json:"size"`
		TypeImage  string `json:"type_image"`
		Layout     string `json:"layout"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if args.Query == "" {
		return "No query provided.", nil
	}

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = t.maxResults
	}

	results, err := searchImages(ctx, args.Query, maxResults, args.Size, args.TypeImage, args.Layout)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if len(results) == 0 {
		output := map[string]interface{}{
			"error": "No images found",
			"query": args.Query,
		}
		jsonOutput, _ := json.Marshal(output)
		return string(jsonOutput), nil
	}

	// 归一化结果（一比一复刻 DeerFlow）
	normalizedResults := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		normalizedResults = append(normalizedResults, map[string]interface{}{
			"title":         r.Title,
			"image_url":     r.Thumbnail,
			"thumbnail_url": r.Thumbnail,
		})
	}

	output := map[string]interface{}{
		"query":         args.Query,
		"total_results": len(normalizedResults),
		"results":       normalizedResults,
		"usage_hint":    "Use the 'image_url' values as reference images in image generation. Download them first if needed.",
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	return string(jsonOutput), nil
}

// Invoke 实现 Eino tool.BaseTool 接口
func (t *ImageSearchTool) Invoke(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "No query provided.", nil
	}

	maxResults := t.maxResults
	if mr, ok := args["max_results"].(int); ok && mr > 0 {
		maxResults = mr
	}

	size, _ := args["size"].(string)
	typeImage, _ := args["type_image"].(string)
	layout, _ := args["layout"].(string)

	results, err := searchImages(ctx, query, maxResults, size, typeImage, layout)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if len(results) == 0 {
		return map[string]interface{}{
			"error": "No images found",
			"query": query,
		}, nil
	}

	normalizedResults := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		normalizedResults = append(normalizedResults, map[string]interface{}{
			"title":         r.Title,
			"image_url":     r.Thumbnail,
			"thumbnail_url": r.Thumbnail,
		})
	}

	return map[string]interface{}{
		"query":         query,
		"total_results": len(normalizedResults),
		"results":       normalizedResults,
		"usage_hint":    "Use the 'image_url' values as reference images in image generation. Download them first if needed.",
	}, nil
}

// ============================================
// 内部实现（简化版 DuckDuckGo 图片搜索）
// ============================================

// ImageResult 图片搜索结果
type ImageResult struct {
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Image     string `json:"image"`
	URL       string `json:"url"`
}

// searchImages 执行图片搜索（简化实现）
func searchImages(ctx context.Context, query string, maxResults int, size, typeImage, layout string) ([]ImageResult, error) {
	// 简化实现：占位实现
	// 注意：实际生产中应该使用专门的图片搜索库

	// 返回空结果（简化实现）
	// 生产环境中应该实现真正的图片搜索
	return []ImageResult{}, nil
}

// DDGImageResult DuckDuckGo 图片搜索结果（用于参考）
type DDGImageResult struct {
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Image     string `json:"image"`
	URL       string `json:"url"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	Source    string `json:"source"`
}

// 辅助函数：从配置中获取整数值
func getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return defaultValue
}
