package logument_test

import (
	"logument"
	"testing"
)

func TestLogument(t *testing.T) {
	// Logument 초기화
	data := map[string]interface{}{"speed": 100, "location": "Seoul"}
	log := logument.NewLogument(data)

	// 초기 스냅샷 출력
	t.Log("Initial Snapshot:", log.CreateSnapshot())

	// Patch 추가
	log.AddPatch("speed", 120)
	log.AddPatch("location", "Busan")
	t.Log("Patches after updates:", log.Patches)

	// 병합 및 결과 출력
	log.MergePatches()
	t.Log("Snapshot after merging patches:", log.CreateSnapshot())
}

func TestLogument_base(t *testing.T) {
	a := make([]int, 10)
	t.Log(a[len(a)-1])
}
