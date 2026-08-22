package service

import (
	"errors"
	"testing"

	"cleanroom-monitor/internal/model"
)

func TestBug04_CompleteErrorChain(t *testing.T) {
	svc := newTestServices(t)
	_, err := svc.Batches.Complete(testCtx(), 999999)
	if err == nil {
		t.Fatal("Complete 不存在批次应返回错误")
	}
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("错误链应能匹配 ErrNotFound，得到: %v", err)
	}
}
