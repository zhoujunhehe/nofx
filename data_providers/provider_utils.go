package data_providers

import (
	"context"
	"fmt"
	"sync"

	"nofx/orchestrator"
)

// FetchFunc 通用的Fetch函数类型
type FetchFunc func(context.Context, *orchestrator.FieldRequest) (interface{}, error)

// FetchBatchConcurrent 并发执行批量获取
func FetchBatchConcurrent(fetchFunc FetchFunc, ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	if len(requests) == 0 {
		return make(map[string]interface{}), nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	result := make(map[string]interface{})

	for _, req := range requests {
		wg.Add(1)
		go func(req *orchestrator.BatchFieldRequest) {
			defer wg.Done()

			value, err := fetchFunc(ctx, req.FieldRequest)

			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = fmt.Errorf("failed to fetch %s: %w", req.Raw, err)
				return
			}
			if err == nil {
				result[req.Raw] = value
			}
		}(req)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return result, nil
}
