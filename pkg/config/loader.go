package config

import (
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

// ============================================
// 全局配置
// ============================================

var (
	globalConfig *AppConfig
	configLoaded bool
)

// ============================================
// ConfigLoader - 配置加载器
// ============================================

// ConfigLoader 配置加载器
type ConfigLoader struct {
	configPath string
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
	}
}

// Load 加载配置
func (l *ConfigLoader) Load() (*AppConfig, error) {
	// 解析配置路径
	resolvedPath, err := ResolveConfigPath(l.configPath)
	if err != nil {
		// 如果找不到配置文件，返回默认配置
		return DefaultAppConfig(), nil
	}

	// 读取文件
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, err
	}

	// 解析 YAML
	var configData map[string]any
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return nil, err
	}

	// 解析环境变量
	resolveEnvVariablesInMap(configData)

	// 加载子配置（保存到全局）
	if titleData, ok := configData["title"]; ok {
		if titleMap, ok := titleData.(map[string]any); ok {
			loadTitleConfigFromMap(titleMap)
		}
	}
	if summarizationData, ok := configData["summarization"]; ok {
		if summarizationMap, ok := summarizationData.(map[string]any); ok {
			loadSummarizationConfigFromMap(summarizationMap)
		}
	}
	if memoryData, ok := configData["memory"]; ok {
		if memoryMap, ok := memoryData.(map[string]any); ok {
			loadMemoryConfigFromMap(memoryMap)
		}
	}
	if subagentsData, ok := configData["subagents"]; ok {
		if subagentsMap, ok := subagentsData.(map[string]any); ok {
			loadSubagentsConfigFromMap(subagentsMap)
		}
	}
	if toolSearchData, ok := configData["tool_search"]; ok {
		if toolSearchMap, ok := toolSearchData.(map[string]any); ok {
			loadToolSearchConfigFromMap(toolSearchMap)
		}
	}
	if checkpointerData, ok := configData["checkpointer"]; ok {
		if checkpointerMap, ok := checkpointerData.(map[string]any); ok {
			loadCheckpointerConfigFromMap(checkpointerMap)
		}
	}

	// 加载 extensions 配置（单独文件）
	extensionsConfig, err := LoadExtensionsConfig()
	if err == nil {
		configData["extensions"] = extensionsConfig
	}

	// 转换为 AppConfig
	config := DefaultAppConfig()
	if err := mapToStruct(configData, config); err != nil {
		return nil, err
	}

	return config, nil
}

// ============================================
// 全局配置函数
// ============================================

// LoadConfig 加载配置（便捷函数）
func LoadConfig(configPath string) (*AppConfig, error) {
	loader := NewConfigLoader(configPath)
	config, err := loader.Load()
	if err != nil {
		return nil, err
	}

	// 保存到全局
	globalConfig = config
	configLoaded = true

	return config, nil
}

// GetConfig 获取全局配置
func GetConfig() *AppConfig {
	if !configLoaded || globalConfig == nil {
		return DefaultAppConfig()
	}
	return globalConfig
}

// SetConfig 设置全局配置
func SetConfig(config *AppConfig) {
	globalConfig = config
	configLoaded = true
}

// ResetConfig 重置全局配置
func ResetConfig() {
	globalConfig = nil
	configLoaded = false
}

// ============================================
// 全局子配置存储
// ============================================

var (
	globalTitleConfig         *TitleConfig
	globalSummarizationConfig *SummarizationConfig
	globalMemoryConfig        *MemoryConfig
	globalSubagentsConfig     *SubagentsConfig
	globalToolSearchConfig    *ToolSearchConfig
	globalCheckpointerConfig  *CheckpointerConfig
)

// ============================================
// 子配置加载函数
// ============================================

func loadTitleConfigFromMap(data map[string]any) {
	config := DefaultTitleConfig()
	if err := mapToStruct(data, &config); err == nil {
		globalTitleConfig = &config
	}
}

func loadSummarizationConfigFromMap(data map[string]any) {
	config := DefaultSummarizationConfig()
	if err := mapToStruct(data, &config); err == nil {
		globalSummarizationConfig = &config
	}
}

func loadMemoryConfigFromMap(data map[string]any) {
	config := DefaultMemoryConfig()
	if err := mapToStruct(data, &config); err == nil {
		globalMemoryConfig = &config
	}
}

func loadSubagentsConfigFromMap(data map[string]any) {
	config := DefaultSubagentsConfig()
	if err := mapToStruct(data, &config); err == nil {
		globalSubagentsConfig = &config
	}
}

func loadToolSearchConfigFromMap(data map[string]any) {
	config := DefaultToolSearchConfig()
	if err := mapToStruct(data, &config); err == nil {
		globalToolSearchConfig = &config
	}
}

func loadCheckpointerConfigFromMap(data map[string]any) {
	config := CheckpointerConfig{}
	if err := mapToStruct(data, &config); err == nil {
		globalCheckpointerConfig = &config
	}
}

// ============================================
// 子配置获取函数
// ============================================

// GetTitleConfig 获取标题配置
func GetTitleConfig() TitleConfig {
	if globalTitleConfig != nil {
		return *globalTitleConfig
	}
	return DefaultTitleConfig()
}

// GetSummarizationConfig 获取摘要配置
func GetSummarizationConfig() SummarizationConfig {
	if globalSummarizationConfig != nil {
		return *globalSummarizationConfig
	}
	return DefaultSummarizationConfig()
}

// GetMemoryConfig 获取记忆配置
func GetMemoryConfig() MemoryConfig {
	if globalMemoryConfig != nil {
		return *globalMemoryConfig
	}
	return DefaultMemoryConfig()
}

// GetSubagentsConfig 获取子代理配置
func GetSubagentsConfig() SubagentsConfig {
	if globalSubagentsConfig != nil {
		return *globalSubagentsConfig
	}
	return DefaultSubagentsConfig()
}

// GetToolSearchConfig 获取工具搜索配置
func GetToolSearchConfig() ToolSearchConfig {
	if globalToolSearchConfig != nil {
		return *globalToolSearchConfig
	}
	return DefaultToolSearchConfig()
}

// GetCheckpointerConfig 获取检查点配置
func GetCheckpointerConfig() *CheckpointerConfig {
	return globalCheckpointerConfig
}

// ============================================
// 环境变量解析辅助函数
// ============================================

func resolveEnvVariablesInMap(data map[string]any) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = ResolveEnvVariable(v)
		case map[string]any:
			resolveEnvVariablesInMap(v)
		case []any:
			resolveEnvVariablesInSlice(v)
		}
	}
}

func resolveEnvVariablesInSlice(data []any) {
	for i, value := range data {
		switch v := value.(type) {
		case string:
			data[i] = ResolveEnvVariable(v)
		case map[string]any:
			resolveEnvVariablesInMap(v)
		case []any:
			resolveEnvVariablesInSlice(v)
		}
	}
}

// ============================================
// map 到 struct 的转换辅助函数
// ============================================

func mapToStruct(data map[string]any, target any) error {
	// 先序列化为 YAML，再反序列化为 struct
	// 这是处理嵌套结构的简单方法
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(yamlData, target)
}

// ============================================
// ExtensionsConfig 加载
// ============================================

const (
	ExtensionsConfigEnvVar  = "DEER_FLOW_EXTENSIONS_CONFIG_PATH"
	ExtensionsConfigFileName = "extensions_config.json"
)

// LoadExtensionsConfig 加载扩展配置
func LoadExtensionsConfig() (*ExtensionsConfig, error) {
	path, err := resolveExtensionsConfigPath("")
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ExtensionsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func resolveExtensionsConfigPath(configPath string) (string, error) {
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
		return "", os.ErrNotExist
	}

	if envPath := os.Getenv(ExtensionsConfigEnvVar); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		return "", os.ErrNotExist
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	localPath := filepath.Join(currentDir, ExtensionsConfigFileName)
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	parentDir := filepath.Dir(currentDir)
	parentPath := filepath.Join(parentDir, ExtensionsConfigFileName)
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath, nil
	}

	return "", os.ErrNotExist
}

// ============================================
// 模型配置查找
// ============================================

// GetModelConfig 获取模型配置
func (c *AppConfig) GetModelConfig(name string) *ModelConfig {
	for i := range c.Models {
		if c.Models[i].Name == name {
			return &c.Models[i]
		}
	}
	return nil
}

// ============================================
// 反射工具（用于后续实现）
// ============================================

// ResolveVariable 解析变量路径（如 "module.path:variable_name"）
// 占位符 - 完整实现需要反射
func ResolveVariable(path string) (any, error) {
	// TODO: 完整实现
	return nil, nil
}

// ResolveClass 解析类路径（如 "module.path:ClassName"）
// 占位符 - 完整实现需要反射
func ResolveClass(path string, baseType reflect.Type) (any, error) {
	// TODO: 完整实现
	return nil, nil
}
