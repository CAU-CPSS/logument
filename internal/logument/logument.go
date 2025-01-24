package logument

import (
	"sync/atomic"
	"time"

	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/CAU-CPSS/logument/internal/jsonpatch"
)

type Ss = jsonr.JsonR
type Pp = jsonpatch.PatchOperation


// Logument 구조체
type Logument struct {
	CurrentSnapshot Snapshot         // 현재 Snapshot
	Snapshots       []Snapshot       // 만들었던 Shanpshot 들의 배열
	Patches         []Patch          // Patch 들의 배열
	Version         []VersionManager // Version 정보 관리
}

type VersionManager struct {
	Version   uint64
	StartTime time.Time
	EndTime   time.Time
}

// NewLogument TODO: initial data 를 json 으로 변경
// Create a new Logument document with the given initial data.
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

// AppendPatch data를 받아 logument에 patches에 patch를 추가
// TODO: Psuedo 구현임 실제 구현 필요
func (l *Logument) AppendPatch(newPatch any) {
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


// Snapshot Snapshot 생성
func (l *Logument) Snapshot(targetVer uint64) Snapshot {
	updatedSnapshot := l.CurrentSnapshot

	return updatedSnapshot
}



// Apply Snapshot과 Patch 병합
func (l *Logument) Apply() {
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

// timestamp 와 version을 같이 관리해야할 것 같음
// 논문 Proposed method 이후에 perspective 혹은 vision이 들어가야할 듯
// 내부적으로는 sqlite 같이 가벼운 DB를 사용해서 구현해보면 좋을 듯
