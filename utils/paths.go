package utils

import (
	"os"
	"path/filepath"
	"sync"
)

const (
	// 向上查找go.mod的最大层数
	maxSearchDepth = 10
)

var (
	promptsDir     string
	promptsDirOnce sync.Once
)

// GetPromptsDir 获取prompts目录的绝对路径
// 查找策略（按优先级）：
// 1. 环境变量 NOFX_PROMPTS_DIR
// 2. 可执行文件所在目录/prompts（生产环境）
// 3. 向上查找go.mod所在目录/prompts（开发环境，最多10层）
// 4. 当前工作目录/prompts（fallback）
func GetPromptsDir() string {
	promptsDirOnce.Do(func() {
		promptsDir = findPromptsDir()
	})
	return promptsDir
}

// findPromptsDir 查找prompts目录
func findPromptsDir() string {
	// 1. 环境变量（最高优先级）
	if envPath := os.Getenv("NOFX_PROMPTS_DIR"); envPath != "" {
		if absPath, err := filepath.Abs(envPath); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}

	// 2. 可执行文件所在目录/prompts（生产环境）
	if execPath, err := os.Executable(); err == nil {
		promptsPath := filepath.Join(filepath.Dir(execPath), "prompts")
		if _, err := os.Stat(promptsPath); err == nil {
			return promptsPath
		}
	}

	// 3. 向上查找go.mod，找到项目根目录/prompts（开发环境）
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < maxSearchDepth; i++ {
			// 检查当前目录的prompts子目录
			promptsPath := filepath.Join(dir, "prompts")
			goModPath := filepath.Join(dir, "go.mod")

			// 如果找到go.mod且prompts存在
			if _, err := os.Stat(goModPath); err == nil {
				if _, err := os.Stat(promptsPath); err == nil {
					return promptsPath
				}
			}

			// 向上一层
			parent := filepath.Dir(dir)
			if parent == dir {
				break // 已到根目录
			}
			dir = parent
		}
	}

	// 4. fallback到当前工作目录/prompts
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, "prompts")
	}

	return "prompts"
}
