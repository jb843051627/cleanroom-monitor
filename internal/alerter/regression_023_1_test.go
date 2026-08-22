package alerter

import (
	"sync"
	"testing"

	"cleanroom-monitor/internal/model"
)

func TestBug23_EvaluatorCloseRace(t *testing.T) {
	e := &Evaluator{queue: make(chan []*model.Reading, 1)}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 10000; i++ {
			e.Submit(nil)
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		e.Close()
	}()
	close(start)
	wg.Wait()
}
