package logging

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type HookExecuter interface {
	Exec(extra map[string]string, b []byte) error
	Close() error
}

type hookOptions struct {
	maxJobs    int
	maxWorkers int
	extra      map[string]string
}

// 设置日志钩子最大缓存数量
func SetHookMaxJobs(maxJobs int) HookOption {
	return func(o *hookOptions) {
		o.maxJobs = maxJobs
	}
}

// 设置日志钩子最大工作线程数量
func SetHookMaxWorkers(maxWorkers int) HookOption {
	return func(o *hookOptions) {
		o.maxWorkers = maxWorkers
	}
}

// 设置日志钩子扩展参数
func SetHookExtra(extra map[string]string) HookOption {
	return func(o *hookOptions) {
		o.extra = extra
	}
}

// HookOption 日志钩子参数选项
type HookOption func(*hookOptions)

// NewHook 创建一个日志钩子
func NewHook(exec HookExecuter, opt ...HookOption) *Hook {
	opts := &hookOptions{
		maxJobs:    1024,
		maxWorkers: 2,
	}

	for _, o := range opt {
		o(opts)
	}

	wg := new(sync.WaitGroup)
	wg.Add(opts.maxWorkers)

	h := &Hook{
		opts: opts,
		q:    make(chan []byte, opts.maxJobs),
		wg:   wg,
		e:    exec,
	}
	h.dispatch()
	return h
}

// Hook 日志钩子将日志写入到数据库
type Hook struct {
	opts   *hookOptions
	q      chan []byte
	wg     *sync.WaitGroup
	e      HookExecuter
	closed int32
}

// dispatch 并发处理日志
func (h *Hook) dispatch() {
	for i := 0; i < h.opts.maxWorkers; i++ {
		go func() {
			defer func() {
				h.wg.Done()
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic in logger hook:", r)
				}
			}()

			for data := range h.q {
				err := h.e.Exec(h.opts.extra, data)
				if err != nil {
					fmt.Println("Failed to write entry:", err.Error())
				}
			}
		}()
	}
}

// Write 日志写入缓冲区
func (h *Hook) Write(p []byte) (int, error) {
	if atomic.LoadInt32(&h.closed) == 1 {
		return len(p), nil
	}
	if len(h.q) == h.opts.maxJobs {
		fmt.Println("Too many jobs, waiting for queue to be empty, discard")
		return len(p), nil
	}

	data := make([]byte, len(p))
	copy(data, p)
	h.q <- data

	return len(p), nil
}

// Flush 刷新日志缓冲区
func (h *Hook) Flush() {
	if atomic.LoadInt32(&h.closed) == 1 {
		return
	}
	atomic.StoreInt32(&h.closed, 1)
	close(h.q)
	h.wg.Wait()
	err := h.e.Close()
	if err != nil {
		fmt.Println("Failed to close logger hook:", err.Error())
	}
}
