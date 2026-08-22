package service

import (
	"context"
	"testing"
)

func TestBug14_RefreshStopsOnCancel(t *testing.T) {
	svc := newTestServices(t)
	ctx, cancel := context.WithCancel(testCtx())
	cancel()
	if err := svc.Dashboard.Refresh(ctx); err == nil {
		t.Fatal("ctx 已取消，Refresh 应返回取消错误")
	}
}
