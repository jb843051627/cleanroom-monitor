package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"cleanroom-monitor/internal/config"
	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/service"
)

// Collector 采集器：接收上报批次入队，后台 worker 批量写入 DB 并触发评估。
type Collector struct {
	cfg      *config.AppConfig
	services *service.Services
	queue    chan *model.ReadingBatch
	closeCh  chan struct{}
	wg       sync.WaitGroup
}

// NewCollector 创建采集器。
func NewCollector(cfg *config.AppConfig, services *service.Services) *Collector {
	return &Collector{
		cfg:      cfg,
		services: services,
		queue:    make(chan *model.ReadingBatch, 2000),
		closeCh:  make(chan struct{}),
	}
}

// Enqueue 上报批次入队（异步，背压丢弃）。
func (c *Collector) Enqueue(batch *model.ReadingBatch) {
	select {
	case c.queue <- batch:
	default:
		log.Printf("[collector] 队列满，丢弃一批读数（%d 条）", len(batch.Readings))
	}
}

// Start 启动采集 worker。
func (c *Collector) Start() {
	c.wg.Add(1)
	go c.run()
}

func (c *Collector) run() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.cfg.CollectorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.closeCh:
			return
		case batch := <-c.queue:
			c.process(batch)
		case <-ticker.C:
			c.flushPending()
		}
	}
}

func (c *Collector) process(batch *model.ReadingBatch) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// 分批截断防止单批过大
	total := len(batch.Readings)
	chunks := c.split(batch.Readings, c.cfg.MaxBatchSize)
	for _, chunk := range chunks {
		n, err := c.services.Readings.Ingest(ctx, &model.ReadingBatch{Readings: chunk})
		if err != nil {
			log.Printf("[collector] 入库失败: %v", err)
			return
		}
		if n > 0 {
			log.Printf("[collector] 入库 %d/%d 条", n, total)
		}
	}
}

func (c *Collector) split(rs []model.ReadingInput, size int) [][]model.ReadingInput {
	if size <= 0 {
		size = 500
	}
	var chunks [][]model.ReadingInput
	for len(rs) > size {
		chunks = append(chunks, rs[:size])
		rs = rs[size:]
	}
	if len(rs) > 0 {
		chunks = append(chunks, rs)
	}
	return chunks
}

// flushPending 兜底处理积压（未来扩展）。
func (c *Collector) flushPending() {
	// 预留：当前队列由 run 主循环消费，无需额外动作
}

// Stop 停止采集器。
func (c *Collector) Stop() {
	close(c.closeCh)
	c.wg.Wait()
}

// QueueLen 当前队列长度。
func (c *Collector) QueueLen() int {
	return len(c.queue)
}