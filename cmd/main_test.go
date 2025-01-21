package main_test

import (
	"testing"

	"github.com/CAU-CPSS/logument/internal/logument"
)

func TestLogument(t *testing.T){
	// Logument 초기화
	data := map[string]interface{}{"speed": 100, "location": "Seoul"}
	lm := logument.NewLogument(data)

	// 초기 스냅샷 출력
	t.Log("Initial Snapshot:", lm)
}