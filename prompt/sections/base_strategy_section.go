package sections

import (
	"fmt"
	"log"
	"nofx/decision"
	"nofx/prompt"
	"nofx/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// BaseStrategySection 基础交易策略部分（从prompts/目录加载）
type BaseStrategySection struct {
	contentName string
	content     string
}

var (
	baseStrategyMap = map[string]*BaseStrategySection{}
	mu              sync.RWMutex
)

// getPromptsDir 获取提示词目录路径
func getPromptsDir() string {
	return utils.GetPromptsDir()
}

func init() {
	// 预加载默认的BaseStrategySection
	promptsDir := getPromptsDir()
	if err := LoadTemplates(promptsDir); err != nil {
		log.Printf("⚠️  加载基础策略提示词模板失败: %v", err)
	} else {
		log.Printf("✓ 已加载 %d 个基础策略提示词模板", len(baseStrategyMap))
	}
}

// SectionName常量
const BaseStrategySectionName = "base_strategy"

// NewBaseStrategySection 创建基础策略Section（从prompts/目录加载默认内容）
func NewBaseStrategySection(contentName string) *BaseStrategySection {
	if contentName == "" {
		contentName = "default"
	}
	mu.RLock()
	defer mu.RUnlock()
	// 使用decision包的GetPromptTemplate加载内容
	baseStrategySection := baseStrategyMap[contentName]
	if baseStrategySection != nil {
		return baseStrategySection
	}

	return &BaseStrategySection{
		contentName: BaseStrategySectionName + "_" + contentName,
		content:     "",
	}
}

func (s *BaseStrategySection) GetName() string {
	return BaseStrategySectionName
}

func (s *BaseStrategySection) GetContent() string {
	return s.content
}

// SetContent 设置自定义内容
func (s *BaseStrategySection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从数据库/文件加载模板
func (s *BaseStrategySection) LoadTemplate(id int64) error {
	// 从prompts/目录重新加载内容
	t, err := decision.GetPromptTemplate(s.contentName)
	if err != nil {
		return err
	}
	if t != nil {
		s.content = t.Content
	}
	return nil
}

func LoadTemplates(dir string) error {
	mu.Lock()
	defer mu.Unlock()

	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("提示词目录不存在: %s", dir)
	}

	// 扫描目录中的所有 .txt 文件
	files, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		return fmt.Errorf("扫描提示词目录失败: %w", err)
	}

	if len(files) == 0 {
		log.Printf("⚠️  提示词目录 %s 中没有找到 .txt 文件", dir)
		return nil
	}

	// 加载每个模板文件
	for _, file := range files {
		// 读取文件内容
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("⚠️  读取提示词文件失败 %s: %v", file, err)
			continue
		}

		// 提取文件名（不含扩展名）作为模板名称
		fileName := filepath.Base(file)
		templateName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// 存储模板
		baseStrategyMap[templateName] = &BaseStrategySection{
			contentName: BaseStrategySectionName + "_" + templateName,
			content:     string(content),
		}

		log.Printf("  📄 加载提示词模板: %s (%s)", templateName, fileName)
	}

	return nil
}
