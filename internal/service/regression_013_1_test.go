package service

import (
	"testing"

	"cleanroom-monitor/internal/model"
)

func TestBug13_RestrictedToReleaseAllowed(t *testing.T) {
	if !model.CanTransition(model.StateRestricted, model.StateRelease) {
		t.Fatal("状态机应允许 restricted → release 转移")
	}
	if !model.CanTransition(model.StateNormal, model.StateRelease) {
		t.Fatal("状态机应允许 normal → release 转移")
	}
}
