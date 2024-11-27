package logument

import (
	"sync/atomic"
	"time"
)

// Logument 구조체
type Logument struct {
	CurrentSnapshot Snapshot // 현재 Snapshot
	Patches         []Patch  // Patch들의 배열
	
}

// NewLogument TODO: initial data 를 json 으로 변경
// Logument 생성
func NewLogument(initialData map[string]any) *Logument {
	return &Logument{
		CurrentSnapshot: Snapshot{
			Version:   1,
			Timestamp: time.Now(),
			Data:      initialData,
		},
		Patches: []Patch{},
	}
}

// CreateSnapshot Snapshot 생성
func (l *Logument) CreateSnapshot(targetVer uint64) Snapshot {
	updatedSnapshot := l.CurrentSnapshot

	return updatedSnapshot
}

// AddPatch Patch 추가
func (l *Logument) AddPatch(newPatch any) {
	switch p := newPatch.(type) {
	case Patch:
		p.VersionCnt = atomic.LoadUint64(&l.CurrentSnapshot.Version) + 1
		l.Patches = append(l.Patches, p)
	case []Patch:
		for i := range p {
			p[i].VersionCnt = atomic.LoadUint64(&l.CurrentSnapshot.Version) + 1
		}
		l.Patches = append(l.Patches, p...)
	default:
		panic("Invalid type for AddPatch. Must be Patch or []Patch.")
	}
}

// MergePatches Snapshot과 Patch 병합
func (l *Logument) MergePatches() {
	for _, patch := range l.Patches {
		switch patch.Op {
		case "replace":
			l.CurrentSnapshot.Data[patch.Path] = patch.Value
		case "add":
			// 예시로 간단히 처리
			if _, exists := l.CurrentSnapshot.Data[patch.Path]; !exists {
				l.CurrentSnapshot.Data[patch.Path] = patch.Value
			}
		case "remove":
			delete(l.CurrentSnapshot.Data, patch.Path)
		}
	}
	l.Patches = []Patch{}
	atomic.AddUint64(&l.CurrentSnapshot.Version, 1)
	l.CurrentSnapshot.Timestamp = time.Now()
}
