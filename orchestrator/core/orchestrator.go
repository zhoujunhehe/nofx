package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	orchestrator "nofx/orchestrator"
)

// DefaultOrchestrator 默认编排器实现（V2版本）
type DefaultOrchestrator struct {
	mu sync.RWMutex

	// Provider管理
	providers map[string]orchestrator.Provider

	// 组件
	parser    *PlaceholderParser
	replacer  *PlaceholderReplacer
	validator *PromptValidator
}

// New 创建新的编排器
func New() orchestrator.Orchestrator {
	return &DefaultOrchestrator{
		providers: make(map[string]orchestrator.Provider),
		parser:    NewPlaceholderParser(),
		replacer:  NewPlaceholderReplacer(),
		validator: NewPromptValidator(),
	}
}

// RegisterProvider 注册Provider
func (o *DefaultOrchestrator) RegisterProvider(provider orchestrator.Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.providers[name]; exists {
		return fmt.Errorf("provider already registered: %s", name)
	}

	o.providers[name] = provider
	return nil
}

// BuildPrompt 构建Prompt
func (o *DefaultOrchestrator) BuildPrompt(ctx context.Context, builder *orchestrator.PromptBuilder) (*orchestrator.BuiltPrompt, error) {
	startTime := time.Now()

	// 1. 合并所有sections用于校验和占位符提取
	allSections := make([]*orchestrator.Section, 0, len(builder.SystemConstraintSections)+len(builder.AnalysisDataSections))
	allSections = append(allSections, builder.SystemConstraintSections...)
	allSections = append(allSections, builder.AnalysisDataSections...)

	// 2. 校验Prompt格式和Provider注册情况
	if err := o.validatePrompt(allSections); err != nil {
		return nil, err
	}

	// 3. 收集所有模板
	templates := make([]string, len(allSections))
	for i, section := range allSections {
		templates[i] = section.Template
	}

	// 4. 提取所有占位符
	placeholders, err := o.parser.ExtractPlaceholders(templates)
	if err != nil {
		return nil, fmt.Errorf("failed to extract placeholders: %w", err)
	}

	// 5. 按Provider分组
	grouped := o.parser.GroupByProvider(placeholders)

	// 6. 并发获取所有数据
	values, err := o.fetchAllData(ctx, grouped)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	// 7. 按照builder指定的顺序构建Prompt
	// SystemConstraintSections数组的顺序就是最终顺序
	systemConstraintPrompt := o.buildPromptFromSections(builder.SystemConstraintSections, placeholders, values)
	// AnalysisDataSections数组的顺序就是最终顺序
	analysisDataPrompt := o.buildPromptFromSections(builder.AnalysisDataSections, placeholders, values)

	return &orchestrator.BuiltPrompt{
		SystemConstraintPrompt: systemConstraintPrompt,
		AnalysisDataPrompt:     analysisDataPrompt,
		Metadata: map[string]interface{}{
			"build_time_ms":      time.Since(startTime).Milliseconds(),
			"placeholders_count": len(placeholders),
			"providers_used":     len(grouped),
		},
	}, nil
}

// fetchAllData 并发获取所有数据
func (o *DefaultOrchestrator) fetchAllData(ctx context.Context, grouped map[string][]*Placeholder) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 错误收集
	errors := make([]error, 0)
	var errMu sync.Mutex

	// 按Provider并发获取
	for providerName, placeholders := range grouped {
		wg.Add(1)
		go func(pName string, phs []*Placeholder) {
			defer wg.Done()

			// 获取Provider
			provider, err := o.getProvider(pName)
			if err != nil {
				errMu.Lock()
				errors = append(errors, fmt.Errorf("provider %s: %w", pName, err))
				errMu.Unlock()
				return
			}

			// 检查Provider是否实现了BatchProvider接口
			if batchProvider, ok := provider.(orchestrator.BatchProvider); ok {
				// 使用批量接口
				batchValues, err := o.fetchBatch(ctx, batchProvider, phs)
				if err != nil {
					errMu.Lock()
					errors = append(errors, fmt.Errorf("batch fetch from %s: %w", pName, err))
					errMu.Unlock()
					return
				}

				// 合并结果
				mu.Lock()
				for k, v := range batchValues {
					values[k] = v
				}
				mu.Unlock()
			} else {
				// 使用单个接口（向后兼容）
				for _, placeholder := range phs {
					request := &orchestrator.FieldRequest{
						Field:  placeholder.Field,
						Params: placeholder.Params,
					}

					value, err := provider.Fetch(ctx, request)
					if err != nil {
						errMu.Lock()
						errors = append(errors, fmt.Errorf("fetch %s from %s: %w", placeholder.Raw, pName, err))
						errMu.Unlock()
						continue
					}

					// 存储值（以Raw占位符为key）
					mu.Lock()
					values[placeholder.Raw] = value
					mu.Unlock()
				}
			}
		}(providerName, placeholders)
	}

	wg.Wait()

	// 检查是否有错误
	if len(errors) > 0 {
		// 返回第一个错误，或者可以返回所有错误
		return values, errors[0]
	}

	return values, nil
}

// fetchBatch 使用批量接口获取数据
func (o *DefaultOrchestrator) fetchBatch(ctx context.Context, provider orchestrator.BatchProvider, placeholders []*Placeholder) (map[string]interface{}, error) {
	// 构建批量请求
	requests := make([]*orchestrator.BatchFieldRequest, len(placeholders))
	for i, placeholder := range placeholders {
		requests[i] = &orchestrator.BatchFieldRequest{
			FieldRequest: &orchestrator.FieldRequest{
				Field:  placeholder.Field,
				Params: placeholder.Params,
			},
			Raw: placeholder.Raw,
		}
	}

	// 调用批量获取
	return provider.FetchBatch(ctx, requests)
}

// getProvider 获取Provider
func (o *DefaultOrchestrator) getProvider(name string) (orchestrator.Provider, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	provider, ok := o.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// validatePrompt 校验Prompt格式和Provider注册情况
func (o *DefaultOrchestrator) validatePrompt(sections []*orchestrator.Section) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// 构建已注册的Provider列表
	registeredProviders := make(map[string]bool)
	for name := range o.providers {
		registeredProviders[name] = true
	}

	// 执行校验
	validationErrors := o.validator.ValidateSections(sections, registeredProviders)

	// 如果有错误，格式化并返回
	if len(validationErrors) > 0 {
		return fmt.Errorf("%s", FormatValidationErrors(validationErrors))
	}

	return nil
}


// buildPromptFromSections 从排序后的sections构建prompt字符串
func (o *DefaultOrchestrator) buildPromptFromSections(sections []*orchestrator.Section, placeholders []*Placeholder, values map[string]interface{}) string {
	parts := make([]string, 0, len(sections))

	for _, section := range sections {
		// 替换占位符
		filled := o.replacer.Replace(section.Template, placeholders, values)
		parts = append(parts, filled)
	}

	// 用两个换行符拼接
	return strings.Join(parts, "\n\n")
}
