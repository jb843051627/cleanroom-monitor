package alerter

import (
	"context"
	"sync"
	"time"

	"cleanroom-monitor/internal/model"
)

// Evaluator 评估 worker 池：串行消费评估队列，避免并发写告警表竞争。
type Evaluator struct {
	engine  *Engine
	queue   chan []*model.Reading
	workers int
	wg      sync.WaitGroup
	closed  bool
}

// NewEvaluator 创建评估 worker 池。
func NewEvaluator(engine *Engine, workers int) *Evaluator {
	if workers <= 0 {
		workers = 1
	}
	e := &Evaluator{
		engine:  engine,
		queue:   make(chan []*model.Reading, 1000),
		workers: workers,
	}
	for i := 0; i < workers; i++ {
		e.wg.Add(1)
		go e.run()
	}
	return e
}

func (e *Evaluator) run() {
	defer e.wg.Done()
	for batch := range e.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = e.engine.Evaluate(ctx, batch)
		cancel()
	}
}

// Submit 提交一批读数进入评估队列。
func (e *Evaluator) Submit(readings []*model.Reading) {
	if e.closed {
		return
	}
	select {
	case e.queue <- readings:
		select {
		case e.queue <- readings:
		default:
		}
	default:
		// 队列满：丢弃（背压保护，防内存膨胀）
	}
}

// Close 关闭 worker 池。
func (e *Evaluator) Close() {
	e.closed = true
	close(e.queue)
	e.wg.Wait()
}